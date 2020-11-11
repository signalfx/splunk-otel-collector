#!/bin/bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/../common.sh

VERSION="${1:-}"
ARCH="${2:-"amd64"}"
OUTPUT_DIR="${3:-"$REPO_DIR/dist/"}"
OTELCOL_REPO_PATH="$REPO_DIR/bin/otelcol_linux_${ARCH}"


if [[ ! -f "$OTELCOL_REPO_PATH" ]]; then
    echo "$OTELCOL_REPO_PATH not found!"
    exit 1
fi

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi

mkdir -p "$OUTPUT_DIR"

fpm -s dir -t deb -n $PKG_NAME -v ${VERSION#v} -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    --architecture "$ARCH" \
    --deb-dist "stable" \
    --deb-user "$SERVICE_USER" \
    --deb-group "$SERVICE_GROUP" \
    --before-install "$PREINSTALL_PATH" \
    --after-install "$POSTINSTALL_PATH" \
    --before-remove "$PREUNINSTALL_PATH" \
    --config-files $SPLUNK_CONFIG_INSTALL_PATH \
    --config-files $OTLP_CONFIG_INSTALL_PATH \
    $SPLUNK_CONFIG_REPO_PATH=$SPLUNK_CONFIG_INSTALL_PATH \
    $OTLP_CONFIG_REPO_PATH=$OTLP_CONFIG_INSTALL_PATH \
    $SPLUNK_ENV_REPO_PATH=$SPLUNK_ENV_INSTALL_PATH \
    $SERVICE_REPO_PATH=$SERVICE_INSTALL_PATH \
    $OTELCOL_REPO_PATH=$OTELCOL_INSTALL_PATH
