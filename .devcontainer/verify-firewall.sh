#!/bin/bash
# Verify the devcontainer firewall configuration.
#
# Read-only: inspects rules and probes connectivity, never modifies the
# firewall.
#
#   verify-firewall.sh                 all checks; exits 1 on any failure
#   verify-firewall.sh --rules-only    rule checks only (sections 1-4)
#   verify-firewall.sh --connectivity  connectivity probes only (section 5)
#   verify-firewall.sh --post-start    rule checks decide the exit code;
#                                      connectivity is reported as advisory
#
# --post-start is what postStartCommand uses: a misconfigured firewall must
# stop the container from coming up, but a transient outage reaching
# npmjs.org must not.
set -uo pipefail

MODE="all"
case "${1:-}" in
    "") MODE="all" ;;
    --rules-only) MODE="rules" ;;
    --connectivity) MODE="connectivity" ;;
    --post-start) MODE="post-start" ;;
    -h | --help)
        sed -n '2,15p' "${BASH_SOURCE[0]}" | sed 's/^# \?//'
        exit 0
        ;;
    *)
        echo "unknown option: $1 (try --help)" >&2
        exit 2
        ;;
esac

RULE_FAIL=0
CONN_FAIL=0
PASS=0
PHASE="rules"

if [ -t 1 ]; then
    C_OK=$'\033[32m'; C_NG=$'\033[31m'; C_HD=$'\033[1m'; C_RS=$'\033[0m'
else
    C_OK=''; C_NG=''; C_HD=''; C_RS=''
fi

ok() { printf '  %sOK%s   %s\n' "$C_OK" "$C_RS" "$1"; PASS=$((PASS + 1)); }
ng() {
    printf '  %sFAIL%s %s\n' "$C_NG" "$C_RS" "$1"
    if [ "$PHASE" = "rules" ]; then
        RULE_FAIL=$((RULE_FAIL + 1))
    else
        CONN_FAIL=$((CONN_FAIL + 1))
    fi
}
section() { printf '\n%s== %s ==%s\n' "$C_HD" "$1" "$C_RS"; }

# iptables needs root; postStartCommand runs this as the remote user, not root.
if [ "$(id -u)" -eq 0 ]; then
    SUDO=""
else
    SUDO="sudo"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

check_commands() {
    section "1. Required commands"
    for cmd in iptables ip6tables ipset dig aggregate jq curl ip; do
        if command -v "$cmd" >/dev/null 2>&1; then
            ok "$cmd is installed"
        else
            ng "$cmd is MISSING (postCreate.sh should install it)"
        fi
    done
}

check_scripts() {
    section "2. Installed firewall scripts"
    for script in init-firewall.sh extend-firewall.sh verify-firewall.sh; do
        path="/usr/local/bin/${script}"
        if [ ! -f "$path" ]; then
            ng "${path} is missing (postCreate.sh did not run?)"
            continue
        fi
        if [ "$(stat -c '%U:%G %a' "$path")" = "root:root 755" ]; then
            ok "${path} is root-owned 0755"
        else
            ng "${path} has unexpected ownership/mode: $(stat -c '%U:%G %a' "$path")"
        fi
    done

    # The installed copies are frozen at postCreate time, so a repo edit only
    # takes effect after a rebuild. Surface that drift rather than hiding it.
    # Only meaningful when running from the repo; from /usr/local/bin there is
    # no repo copy to compare against.
    if [ "$SCRIPT_DIR" = "/usr/local/bin" ]; then
        ok "running from /usr/local/bin (repo drift check skipped)"
        return
    fi
    for script in extend-firewall.sh verify-firewall.sh; do
        if [ -f "${SCRIPT_DIR}/${script}" ] && [ -f "/usr/local/bin/${script}" ]; then
            if diff -q "${SCRIPT_DIR}/${script}" "/usr/local/bin/${script}" >/dev/null 2>&1; then
                ok "installed ${script} matches the repo copy"
            else
                ng "installed ${script} differs from the repo copy (rebuild to apply)"
            fi
        fi
    done
}

check_ipv4() {
    section "3. IPv4 rules"
    local output_rules
    output_rules="$($SUDO iptables -S OUTPUT 2>/dev/null)"

    if grep -q '^-P OUTPUT DROP$' <<<"$output_rules"; then
        ok "OUTPUT default policy is DROP"
    else
        ng "OUTPUT default policy is not DROP"
    fi

    # A blanket exception has no -d, so it applies to every destination.
    if grep -qE '^-A OUTPUT -p tcp .*--dport 22 -j ACCEPT$' <<<"$output_rules"; then
        ng "blanket outbound SSH (tcp/22 to anywhere) is still allowed"
    else
        ok "no blanket outbound SSH exception"
    fi

    if grep -qE '^-A OUTPUT -p udp .*--dport 53 -j ACCEPT$' <<<"$output_rules"; then
        ng "blanket outbound DNS (udp/53 to anywhere) is still allowed"
    else
        ok "no blanket outbound DNS exception"
    fi

    local dns_servers
    mapfile -t dns_servers < <(awk '/^nameserver/ {print $2}' /etc/resolv.conf)
    if [ "${#dns_servers[@]}" -eq 0 ]; then
        ng "no nameserver found in /etc/resolv.conf"
    fi
    for dns in "${dns_servers[@]}"; do
        if grep -qE "^-A OUTPUT -d ${dns}(/32)? -p udp .*--dport 53 -j ACCEPT$" <<<"$output_rules"; then
            ok "DNS is allowed to resolver ${dns} only"
        else
            ng "no DNS allow rule for resolver ${dns} (DNS resolution will break)"
        fi
    done

    if grep -q 'match-set allowed-domains dst -j ACCEPT' <<<"$output_rules"; then
        ok "allowed-domains allowlist rule is present"
    else
        ng "allowed-domains allowlist rule is missing"
    fi

    if grep -q '^-A OUTPUT -j REJECT' <<<"$output_rules"; then
        ok "catch-all REJECT rule is present"
    else
        ng "catch-all REJECT rule is missing"
    fi
}

check_ipv6() {
    section "4. IPv6 rules"
    local ip6_rules
    ip6_rules="$($SUDO ip6tables -S 2>/dev/null)"
    for chain in INPUT FORWARD OUTPUT; do
        if grep -q "^-P ${chain} DROP$" <<<"$ip6_rules"; then
            ok "ip6tables ${chain} policy is DROP"
        else
            ng "ip6tables ${chain} policy is not DROP (IPv6 would bypass the allowlist)"
        fi
    done
    if grep -q '^-A INPUT -i lo -j ACCEPT$' <<<"$ip6_rules" &&
        grep -q '^-A OUTPUT -o lo -j ACCEPT$' <<<"$ip6_rules"; then
        ok "IPv6 loopback is still allowed"
    else
        ng "IPv6 loopback is not allowed (localhost resolving to ::1 will break)"
    fi
}

check_connectivity() {
    section "5. Actual connectivity"
    if [ -n "$(dig +short A github.com 2>/dev/null)" ]; then
        ok "DNS resolution works"
    else
        ng "DNS resolution failed"
    fi

    for domain in api.github.com registry.npmjs.org proxy.golang.org codeload.github.com; do
        if curl -s --connect-timeout 5 --max-time 10 -o /dev/null "https://${domain}"; then
            ok "reachable: ${domain}"
        else
            ng "NOT reachable: ${domain} (should be on the allowlist)"
        fi
    done

    for domain in example.com www.google.com; do
        if curl -s --connect-timeout 5 --max-time 10 -o /dev/null "https://${domain}"; then
            ng "reachable: ${domain} (should be blocked)"
        else
            ok "blocked: ${domain}"
        fi
    done

    # DNS to an off-allowlist resolver is the tunneling path the firewall closes.
    if timeout 8 dig @8.8.8.8 example.com +time=3 +tries=1 +short >/dev/null 2>&1; then
        ng "external resolver 8.8.8.8 is reachable (DNS tunneling is possible)"
    else
        ok "external resolver 8.8.8.8 is blocked"
    fi

    # SSH must still work to GitHub: package-lock.json has a git+ssh dependency.
    if timeout 8 bash -c 'exec 3<>/dev/tcp/github.com/22' 2>/dev/null; then
        ok "SSH to github.com:22 works (git+ssh dependencies can install)"
    else
        ng "SSH to github.com:22 failed (git+ssh dependencies will not install)"
    fi
    if timeout 8 bash -c 'exec 3<>/dev/tcp/gitlab.com/22' 2>/dev/null; then
        ng "SSH to gitlab.com:22 works (should be blocked)"
    else
        ok "SSH to an off-allowlist host is blocked"
    fi

    # Port 9 on ::1 is closed, so an allowed packet is refused immediately
    # (curl exit 7) while a dropped one times out (exit 28). --noproxy keeps a
    # proxy setting from turning this into a probe of the proxy instead. Any
    # other exit code means the probe never established what it set out to, so
    # report it rather than passing. Do not use ping6 here: it does not exist
    # on Ubuntu 24.04 and its absence looks like a firewall block.
    local rc
    curl -s --noproxy '*' --connect-timeout 3 -o /dev/null "http://[::1]:9/" 2>/dev/null
    rc=$?
    case "$rc" in
        0 | 7) ok "IPv6 loopback reachable (connection refused, not dropped)" ;;
        28) ng "IPv6 loopback timed out (localhost via ::1 is blocked)" ;;
        *) ng "IPv6 loopback probe inconclusive (curl exit ${rc})" ;;
    esac
}

run_rule_checks() {
    PHASE="rules"
    check_commands
    check_scripts
    check_ipv4
    check_ipv6
}

run_connectivity_checks() {
    PHASE="connectivity"
    check_connectivity
}

case "$MODE" in
    all | post-start)
        run_rule_checks
        run_connectivity_checks
        ;;
    rules) run_rule_checks ;;
    connectivity) run_connectivity_checks ;;
esac

printf '\n%s== Summary ==%s\n' "$C_HD" "$C_RS"
printf '  passed: %d, rule failures: %d, connectivity failures: %d\n' \
    "$PASS" "$RULE_FAIL" "$CONN_FAIL"

if [ "$MODE" = "post-start" ]; then
    # Connectivity is advisory here: an external outage must not stop the
    # container from starting. Rule failures are configuration errors and do.
    if [ "$CONN_FAIL" -gt 0 ]; then
        printf '  %sWARNING%s: %d connectivity check(s) failed (not blocking startup)\n' \
            "$C_NG" "$C_RS" "$CONN_FAIL"
    fi
    [ "$RULE_FAIL" -eq 0 ] || exit 1
else
    [ "$((RULE_FAIL + CONN_FAIL))" -eq 0 ] || exit 1
fi
