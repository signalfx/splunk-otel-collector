#!/bin/bash

# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
TRANSLATESFX_INSTALL_PATH="/usr/bin/translatesfx"
AGENT_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/agent_config.yaml"
AGENT_CONFIG_INSTALL_PATH="/etc/otel/collector/agent_config.yaml"
GATEWAY_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/gateway_config.yaml"
GATEWAY_CONFIG_INSTALL_PATH="/etc/otel/collector/gateway_config.yaml"
SERVICE_REPO_PATH="$FPM_DIR/$SERVICE_NAME.service"
SERVICE_INSTALL_PATH="/lib/systemd/system/$SERVICE_NAME.service"

FLUENTD_CONFIG_INSTALL_DIR="/etc/otel/collector/fluentd"

SMART_AGENT_RELEASE_PATH="${FPM_DIR}/../smart-agent-release.txt"
JMX_LIB_VERSION_PATH="${FPM_DIR}/../jmx-lib-version.txt"
SMART_AGENT_RELEASE_URL="https://api.github.com/repos/signalfx/signalfx-agent/releases"
BUNDLE_BASE_DIR="/usr/lib/splunk-otel-collector"
AGENT_BUNDLE_INSTALL_DIR="$BUNDLE_BASE_DIR/agent-bundle"

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

download_smart_agent() {
    local tag="$1"
    local buildroot="$2"
    local api_url=""
    local dl_url=""

    if [ "$tag" = "latest" ]; then
        tag=$( curl -sL "$SMART_AGENT_RELEASE_URL/latest" | jq -r '.tag_name' )
        if [ -z "$tag" ]; then
            echo "Failed to get tag_name for latest release from $SMART_AGENT_RELEASE_URL/latest" >&2
            exit 1
        fi
    fi

    api_url="$SMART_AGENT_RELEASE_URL/tags/$tag"
    dl_url="$( curl -sL "$api_url" | jq -r '.assets[] .browser_download_url' | grep "signalfx-agent-${tag#v}\.tar\.gz" )"
    if [ -z "$dl_url" ]; then
        echo "Failed to get the agent download url from $api_url" >&2
        exit 1
    fi

    echo "Downloading $dl_url ..."
    curl -sL "$dl_url" -o "$buildroot/signalfx-agent.tar.gz"

    mkdir -p "$buildroot/$BUNDLE_BASE_DIR"
    tar -xzf "$buildroot/signalfx-agent.tar.gz" -C "$buildroot/$BUNDLE_BASE_DIR"
    mv "$buildroot/$BUNDLE_BASE_DIR/signalfx-agent" "$buildroot/$AGENT_BUNDLE_INSTALL_DIR"
    find "$buildroot/$AGENT_BUNDLE_INSTALL_DIR" -wholename "*test*.key" -delete -or -wholename "*test*.pem" -delete
    rm -f "$buildroot/$AGENT_BUNDLE_INSTALL_DIR/bin/signalfx-agent"
    rm -f "$buildroot/$AGENT_BUNDLE_INSTALL_DIR/bin/agent-status"
    rm -f "$buildroot/signalfx-agent.tar.gz"
}

download_and_jmx_jar() {
    local version="$1"
    local buildroot="$2"

    JMX_LIB_RELEASE_DL_URL="https://repo1.maven.org/maven2/io/opentelemetry/contrib/opentelemetry-jmx-metrics/$version/opentelemetry-jmx-metrics-$version.jar"
    mkdir -p "$buildroot/opt"

    echo "Downloading ${JMX_LIB_RELEASE_DL_URL}"
    curl -sL "$JMX_LIB_RELEASE_DL_URL" -o "$buildroot/opt/opentelemetry-java-contrib-jmx-metrics.jar"
}

setup_files_and_permissions() {
    local otelcol="$1"
    local translatesfx="$2"
    local buildroot="$3"

    create_user_group

    mkdir -p "$buildroot/$(dirname $OTELCOL_INSTALL_PATH)"
    cp -f "$otelcol" "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chown root:root "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$OTELCOL_INSTALL_PATH"

    mkdir -p "$buildroot/$(dirname $TRANSLATESFX_INSTALL_PATH)"
    cp -f "$translatesfx" "$buildroot/$TRANSLATESFX_INSTALL_PATH"
    sudo chown root:root "$buildroot/$TRANSLATESFX_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$TRANSLATESFX_INSTALL_PATH"

    cp -r "$FPM_DIR/etc" "$buildroot/etc"
    cp -f "$AGENT_CONFIG_REPO_PATH" "$buildroot/$AGENT_CONFIG_INSTALL_PATH"
    cp -f "$GATEWAY_CONFIG_REPO_PATH" "$buildroot/$GATEWAY_CONFIG_INSTALL_PATH"
    sudo chown -R $SERVICE_USER:$SERVICE_GROUP "$buildroot/etc/otel"
    sudo chmod -R 755 "$buildroot/etc/otel"
    sudo chmod 600 "$buildroot/etc/otel/collector/$SERVICE_NAME.conf.example"

    mkdir -p "$buildroot/$(dirname $SERVICE_INSTALL_PATH)"
    cp -f "$SERVICE_REPO_PATH" "$buildroot/$SERVICE_INSTALL_PATH"
    sudo chown root:root "$buildroot/$SERVICE_INSTALL_PATH"
    sudo chmod 644 "$buildroot/$SERVICE_INSTALL_PATH"

    if [ -d "$buildroot/$BUNDLE_BASE_DIR" ]; then
        sudo chown -R $SERVICE_USER:$SERVICE_GROUP "$buildroot/$BUNDLE_BASE_DIR"
        sudo chmod -R 755 "$buildroot/$BUNDLE_BASE_DIR"
    fi
}
