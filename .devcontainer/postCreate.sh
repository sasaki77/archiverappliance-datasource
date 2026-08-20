#!/bin/bash
set -e

echo "=== Installing firewall script ==="
INIT_FIREWALL_COMMIT="d945a61bc6346abce607252d5667df8a3bf0461a"
INIT_FIREWALL_SHA256="d6da27a11b17bd9b1effa1d8d695e8e223c2214e85949f3afc6b6381d5327bae"
curl -fsSL "https://raw.githubusercontent.com/anthropics/claude-code/${INIT_FIREWALL_COMMIT}/.devcontainer/init-firewall.sh" -o /tmp/init-firewall.sh
echo "${INIT_FIREWALL_SHA256}  /tmp/init-firewall.sh" | sha256sum -c -
sudo mv /tmp/init-firewall.sh /usr/local/bin/init-firewall.sh
sudo chmod +x /usr/local/bin/init-firewall.sh

echo "=== Change .claude owner ==="
sudo chown -R vscode:vscode /home/vscode/.claude

echo "=== Installing project dependencies ==="
go install github.com/magefile/mage@latest
