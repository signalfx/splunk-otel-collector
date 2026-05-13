#!/usr/bin/env bash
set -euo pipefail

# Installation instructions: https://github.com/mikefarah/yq?tab=readme-ov-file#install

YQ_VERSION="v4.44.3"
YQ_BINARY="yq_linux_amd64"

if command -v yq &> /dev/null; then
  echo "yq is already installed: $(yq --version)"
  exit 0
fi

echo "Installing yq ${YQ_VERSION}..."
wget -qO /usr/bin/yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/${YQ_BINARY}"
chmod +x /usr/bin/yq

echo "yq installed: $(yq --version)"
