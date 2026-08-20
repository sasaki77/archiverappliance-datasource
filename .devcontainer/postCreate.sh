#!/bin/bash
set -e

echo "=== Installing firewall script ==="
sudo curl -fsSL https://raw.githubusercontent.com/anthropics/claude-code/main/.devcontainer/init-firewall.sh -o /usr/local/bin/init-firewall.sh
sudo chmod +x /usr/local/bin/init-firewall.sh

echo "=== Change .claude owner ==="
sudo chown -R vscode:vscode /home/vscode/.claude

echo "=== Installing project dependencies ==="
go install github.com/magefile/mage@latest
