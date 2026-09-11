#!/bin/bash
set -e

echo "=== Installing firewall dependencies ==="
# init-firewall.sh / extend-firewall.sh need these; they are not in the base
# image. Install here (postCreate) while the network is still unrestricted —
# once postStartCommand brings the firewall up, apt archives are unreachable.
sudo apt-get update
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    iptables ipset dnsutils aggregate jq

echo "=== Installing firewall script ==="
INIT_FIREWALL_COMMIT="d945a61bc6346abce607252d5667df8a3bf0461a"
INIT_FIREWALL_SHA256="d6da27a11b17bd9b1effa1d8d695e8e223c2214e85949f3afc6b6381d5327bae"
INIT_FIREWALL_TMP="$(mktemp)"
trap 'rm -f "${INIT_FIREWALL_TMP}"' EXIT
curl -fsSL --max-time 30 "https://raw.githubusercontent.com/anthropics/claude-code/${INIT_FIREWALL_COMMIT}/.devcontainer/init-firewall.sh" -o "${INIT_FIREWALL_TMP}"
echo "${INIT_FIREWALL_SHA256}  ${INIT_FIREWALL_TMP}" | sha256sum -c -
sudo install -m 0755 -o root -g root "${INIT_FIREWALL_TMP}" /usr/local/bin/init-firewall.sh

echo "=== Installing firewall extension script ==="
sudo install -m 0755 -o root -g root "$(dirname "${BASH_SOURCE[0]}")/extend-firewall.sh" /usr/local/bin/extend-firewall.sh

echo "=== Installing firewall verification script ==="
sudo install -m 0755 -o root -g root "$(dirname "${BASH_SOURCE[0]}")/verify-firewall.sh" /usr/local/bin/verify-firewall.sh

echo "=== Change .claude owner ==="
sudo chown -R vscode:vscode /home/vscode/.claude

echo "=== Installing project dependencies ==="
go install github.com/magefile/mage@latest
