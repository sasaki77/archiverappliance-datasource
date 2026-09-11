#!/bin/bash
set -euo pipefail

# init-firewall.sh unconditionally allows outbound TCP/22 and UDP/53 to any
# destination, which bypasses the allowed-domains allowlist for those ports.
# Narrow DNS to the resolvers this container actually uses and drop the blanket
# SSH exception; SSH to hosts already in allowed-domains (e.g. GitHub, needed
# for a git+ssh npm dependency) still works via the generic allowed-domains
# match init-firewall.sh appends for all protocols/ports.
mapfile -t DNS_SERVERS < <(awk '/^nameserver/ {print $2}' /etc/resolv.conf)
if [ "${#DNS_SERVERS[@]}" -eq 0 ]; then
    echo "ERROR: No nameserver found in /etc/resolv.conf" >&2
    exit 1
fi

iptables -D OUTPUT -p udp --dport 53 -j ACCEPT 2>/dev/null || true
iptables -D INPUT -p udp --sport 53 -j ACCEPT 2>/dev/null || true
iptables -D OUTPUT -p tcp --dport 22 -j ACCEPT 2>/dev/null || true
iptables -D INPUT -p tcp --sport 22 -m state --state ESTABLISHED -j ACCEPT 2>/dev/null || true

for dns in "${DNS_SERVERS[@]}"; do
    echo "Allowing DNS to resolver ${dns}"
    iptables -I OUTPUT 1 -p udp --dport 53 -d "$dns" -j ACCEPT
    iptables -I INPUT 1 -p udp --sport 53 -s "$dns" -j ACCEPT
done

# init-firewall.sh only manages IPv4, leaving every ip6tables policy at ACCEPT.
# The allowlist is IPv4-only by design (A records into an IPv4 ipset), so rather
# than mirroring it for AAAA, deny IPv6 outright — otherwise any future IPv6
# connectivity would bypass allowed-domains entirely. Loopback is kept so that
# `localhost` resolving to ::1 (local Grafana, e2e runs) still works.
echo "Denying non-loopback IPv6..."
ip6tables -F
ip6tables -A INPUT -i lo -j ACCEPT
ip6tables -A OUTPUT -o lo -j ACCEPT
ip6tables -P INPUT DROP
ip6tables -P FORWARD DROP
ip6tables -P OUTPUT DROP

for chain in INPUT FORWARD OUTPUT; do
    policy=$(ip6tables -S "$chain" | awk 'NR==1 {print $3}')
    if [ "$policy" != "DROP" ]; then
        echo "ERROR: ip6tables ${chain} policy is ${policy}, expected DROP" >&2
        exit 1
    fi
    echo "OK: ip6tables ${chain} policy is DROP"
done

ALLOWED_DOMAINS=(
    proxy.golang.org
    sum.golang.org
    storage.googleapis.com
    repo.yarnpkg.com
    registry.yarnpkg.com
    codeload.github.com
)

for domain in "${ALLOWED_DOMAINS[@]}"; do
    for ip in $(dig +short A "$domain"); do
        if ! ipset add allowed-domains "$ip" -exist 2>/dev/null; then
            echo "ERROR: Failed to add ${ip} (resolved from ${domain}) to allowed-domains ipset" >&2
            exit 1
        fi
    done
done

echo "Verifying Go proxy access..."
for domain in "${ALLOWED_DOMAINS[@]}"; do
    if ! curl --connect-timeout 5 --max-time 10 -s -o /dev/null "https://${domain}"; then
        echo "ERROR: Firewall verification failed - unable to reach https://${domain}"
        exit 1
    fi
    echo "OK: https://${domain} is reachable"
done
