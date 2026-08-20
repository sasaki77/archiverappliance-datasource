#!/bin/bash
set -euo pipefail

ALLOWED_DOMAINS=(
    proxy.golang.org
    sum.golang.org
    storage.googleapis.com
    repo.yarnpkg.com
    registry.yarnpkg.com
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
