#!/bin/bash

FPM_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$FPM_DIR/../../../../" && pwd )"

PKG_NAME="splunk-otel-collector"
PKG_VENDOR="Splunk, Inc."
PKG_MAINTAINER="Splunk, Inc."
PKG_DESCRIPTION="Splunk OpenTelemetry Collector"
PKG_LICENSE="Apache 2.0"
PKG_URL="https://github.com/signalfx/splunk-otel-collector"

SERVICE_NAME="splunk-otel-collector"
SERVICE_USER="splunk-otel-collector"
SERVICE_GROUP="splunk-otel-collector"

OTELCOL_INSTALL_PATH="/usr/bin/otelcol"
CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/agent_config.yaml"
CONFIG_INSTALL_PATH="/etc/otel/collector/gateway_config"
SERVICE_REPO_PATH="$FPM_DIR/$SERVICE_NAME.service"
SERVICE_INSTALL_PATH="/lib/systemd/system/$SERVICE_NAME.service"

PREINSTALL_PATH="$FPM_DIR/preinstall.sh"
POSTINSTALL_PATH="$FPM_DIR/postinstall.sh"
PREUNINSTALL_PATH="$FPM_DIR/preuninstall.sh"

get_version() {
    commit_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true )"
    if [[ -z "$commit_tag" ]]; then
        latest_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true )"
        if [[ -n "$latest_tag" ]]; then
            echo "${latest_tag}-post"
        else
            echo "0.0.1-post"
        fi
    else
        echo "$commit_tag"
    fi
}

create_user_group() {
    sudo getent passwd $SERVICE_USER >/dev/null || \
        sudo useradd --system --user-group --no-create-home --shell /sbin/nologin $SERVICE_USER
}

setup_files_and_permissions() {
    local otelcol="$1"
    local buildroot="$2"

    create_user_group

    mkdir -p "$buildroot/$(dirname $OTELCOL_INSTALL_PATH)"
    cp -f "$otelcol" "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chown root:root "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$OTELCOL_INSTALL_PATH"

    cp -r "$FPM_DIR/etc" "$buildroot/etc"
    cp -f "$CONFIG_REPO_PATH" "$buildroot/$CONFIG_INSTALL_PATH"
    sudo chown -R $SERVICE_USER:$SERVICE_GROUP "$buildroot/etc/otel"
    sudo chmod -R 755 "$buildroot/etc/otel"
    sudo chmod 600 "$buildroot/etc/otel/collector/splunk_env.example"

    mkdir -p "$buildroot/$(dirname $SERVICE_INSTALL_PATH)"
    cp -f "$SERVICE_REPO_PATH" "$buildroot/$SERVICE_INSTALL_PATH"
    sudo chown root:root "$buildroot/$SERVICE_INSTALL_PATH"
    sudo chmod 644 "$buildroot/$SERVICE_INSTALL_PATH"
}
