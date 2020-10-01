#!/bin/bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/../../../../../" && pwd )"
VERSION="${1:-}"
ARCH="${2:-"amd64"}"
OUTPUT_DIR="${3:-"$REPO_DIR/dist/"}"
OTELCOL_PATH="$REPO_DIR/bin/otelcol_linux_$ARCH"
CONFIG_PATH="$REPO_DIR/cmd/otelcol/config/collector/splunk_config.yaml"
ENV_PATH="$REPO_DIR/internal/buildscripts/packaging/fpm/splunk-env.conf"

. $SCRIPT_DIR/../common.sh

if [[ -z "$VERSION" ]]; then
    latest_tag="$( git describe --abbrev=0 --match v[0-9]* 2>/dev/null || true )"
    if [[ -n "$latest_tag" ]]; then
        VERSION="${latest_tag}~post"
    else
        VERSION="0.0.1~post"
    fi
fi

mkdir -p "$OUTPUT_DIR"

fpm -s dir -t rpm -n $PKG_NAME -v ${VERSION#v} -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    --architecture "$ARCH" \
    --rpm-summary "$PKG_DESCRIPTION" \
    --rpm-user "$SERVICE_USER" \
    --rpm-group "$SERVICE_GROUP" \
    --before-install "$PREINSTALL_PATH" \
    --after-install "$POSTINSTALL_PATH" \
    --pre-uninstall "$PREUNINSTALL_PATH" \
    --config-files /etc/splunk-otel-collector/config.yaml \
    --config-files /etc/systemd/system/splunk-otel-collector.service.d/splunk-env.conf \
    $CONFIG_PATH=/etc/splunk-otel-collector/config.yaml \
    $ENV_PATH=/etc/systemd/system/splunk-otel-collector.service.d/splunk-env.conf \
    $SERVICE_PATH=/lib/systemd/system/$SERVICE_NAME.service \
    $OTELCOL_PATH=/usr/bin/otelcol
