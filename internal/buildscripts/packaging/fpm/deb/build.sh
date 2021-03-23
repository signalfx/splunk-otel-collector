#!/bin/bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/../common.sh

VERSION="${1:-}"
ARCH="${2:-amd64}"
OUTPUT_DIR="${3:-$REPO_DIR/dist}"
SMART_AGENT_RELEASE="${4:-latest}"

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi
VERSION="${VERSION#v}"

otelcol_path="$REPO_DIR/bin/otelcol_linux_${ARCH}"

buildroot="$(mktemp -d)"

if [ "$ARCH" = "amd64" ]; then
    download_smart_agent "$SMART_AGENT_RELEASE" "$buildroot"
fi

setup_files_and_permissions "$otelcol_path" "$buildroot"

mkdir -p "$OUTPUT_DIR"

sudo fpm -s dir -t deb -n "$PKG_NAME" -v "$VERSION" -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    --architecture "$ARCH" \
    --deb-dist "stable" \
    --deb-use-file-permissions \
    --before-install "$PREINSTALL_PATH" \
    --after-install "$POSTINSTALL_PATH" \
    --before-remove "$PREUNINSTALL_PATH" \
    --deb-no-default-config-files \
    --config-files "$CONFIG_INSTALL_PATH" \
    --config-files "$FLUENTD_CONFIG_INSTALL_DIR" \
    "$buildroot/"=/

dpkg -c "${OUTPUT_DIR}/${PKG_NAME}_${VERSION}_${ARCH}.deb"
