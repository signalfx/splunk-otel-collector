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
REPO_DIR="$( cd "$FPM_DIR/../../" && pwd )"

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
LAUNCHER_INSTALL_PATH="/usr/bin/splunk-otel-collector-launcher"
OPAMPSUPERVISOR_INSTALL_PATH="/usr/bin/opampsupervisor"
CONFIG_DIR_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/config.d.linux"
CONFIG_DIR_INSTALL_PATH="/etc/otel/collector/config.d"
AGENT_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/agent_config.yaml"
AGENT_CONFIG_INSTALL_PATH="/etc/otel/collector/agent_config.yaml"
GATEWAY_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/gateway_config.yaml"
GATEWAY_CONFIG_INSTALL_PATH="/etc/otel/collector/gateway_config.yaml"
LOGS_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/splunk_logs_config_linux.yaml"
LOGS_CONFIG_INSTALL_PATH="/etc/otel/collector/splunk_logs_config_linux.yaml"
METRICS_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/splunk_metrics_config_linux.yaml"
METRICS_CONFIG_INSTALL_PATH="/etc/otel/collector/splunk_metrics_config_linux.yaml"
SUPERVISOR_CONFIG_DIR_INSTALL_PATH="/etc/otel/collector/supervisor"
SUPERVISOR_STORAGE_PARENT_INSTALL_PATH="/var/lib/otelcol"
SERVICE_REPO_PATH="$FPM_DIR/$SERVICE_NAME.service"
SERVICE_INSTALL_PATH="/lib/systemd/system/$SERVICE_NAME.service"

JMX_METRIC_GATHERER_RELEASE_PATH="${FPM_DIR}/../jmx-metric-gatherer-release.txt"

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

download_jmx_metric_gatherer() {
    local version="$1"
    local buildroot="$2"

    JMX_METRIC_GATHERER_RELEASE_DL_URL="https://github.com/open-telemetry/opentelemetry-java-contrib/releases/download/$version/opentelemetry-jmx-metrics.jar"
    mkdir -p "$buildroot/opt"

    echo "Downloading ${JMX_METRIC_GATHERER_RELEASE_DL_URL}"
    curl -sL "$JMX_METRIC_GATHERER_RELEASE_DL_URL" -o "$buildroot/opt/opentelemetry-java-contrib-jmx-metrics.jar"
}

setup_files_and_permissions() {
    local otelcol="$1"
    local buildroot="$2"
    local launcher="${3:-}"
    local opampsupervisor="${4:-}"
    local with_opamp_supervisor="${WITH_OPAMP_SUPERVISOR:-false}"

    create_user_group

    mkdir -p "$buildroot/$(dirname $OTELCOL_INSTALL_PATH)"
    cp -f "$otelcol" "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chown root:root "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$OTELCOL_INSTALL_PATH"

    # Building with the supervisor flag will add the launcher and OpAMP Supervisor binaries,
    # create necessary dirs, and update the service entrypoint to use the launcher.
    if [[ "$with_opamp_supervisor" = "true" ]]; then
        if [[ ! -f "$launcher" ]]; then
            echo "Launcher binary not found: $launcher" >&2
            exit 1
        fi
        if [[ ! -f "$opampsupervisor" ]]; then
            echo "OpAMP Supervisor binary not found: $opampsupervisor" >&2
            exit 1
        fi

        cp -f "$launcher" "$buildroot/$LAUNCHER_INSTALL_PATH"
        sudo chown root:root "$buildroot/$LAUNCHER_INSTALL_PATH"
        sudo chmod 755 "$buildroot/$LAUNCHER_INSTALL_PATH"

        cp -f "$opampsupervisor" "$buildroot/$OPAMPSUPERVISOR_INSTALL_PATH"
        sudo chown root:root "$buildroot/$OPAMPSUPERVISOR_INSTALL_PATH"
        sudo chmod 755 "$buildroot/$OPAMPSUPERVISOR_INSTALL_PATH"

        mkdir -p "$buildroot/$SUPERVISOR_STORAGE_PARENT_INSTALL_PATH"
        sudo chown $SERVICE_USER:$SERVICE_GROUP "$buildroot/$SUPERVISOR_STORAGE_PARENT_INSTALL_PATH"
        sudo chmod 755 "$buildroot/$SUPERVISOR_STORAGE_PARENT_INSTALL_PATH"
    fi

    cp -r "$FPM_DIR/etc" "$buildroot/etc"
    cp -r "$CONFIG_DIR_REPO_PATH" "$buildroot/$CONFIG_DIR_INSTALL_PATH"
    cp -f "$AGENT_CONFIG_REPO_PATH" "$buildroot/$AGENT_CONFIG_INSTALL_PATH"
    cp -f "$GATEWAY_CONFIG_REPO_PATH" "$buildroot/$GATEWAY_CONFIG_INSTALL_PATH"
    cp -f "$LOGS_CONFIG_REPO_PATH" "$buildroot/$LOGS_CONFIG_INSTALL_PATH"
    cp -f "$METRICS_CONFIG_REPO_PATH" "$buildroot/$METRICS_CONFIG_INSTALL_PATH"
    if [[ "$with_opamp_supervisor" = "true" ]]; then
        mkdir -p "$buildroot/$SUPERVISOR_CONFIG_DIR_INSTALL_PATH"
    fi
    sudo chown -R $SERVICE_USER:$SERVICE_GROUP "$buildroot/etc/otel"
    sudo chmod -R 755 "$buildroot/etc/otel"
    sudo chmod 600 "$buildroot/etc/otel/collector/$SERVICE_NAME.conf.example"

    mkdir -p "$buildroot/$(dirname $SERVICE_INSTALL_PATH)"
    cp -f "$SERVICE_REPO_PATH" "$buildroot/$SERVICE_INSTALL_PATH"
    if [[ "$with_opamp_supervisor" = "true" ]]; then
        sed -i 's#^ExecStart=/usr/bin/otelcol #ExecStart=/usr/bin/splunk-otel-collector-launcher #' "$buildroot/$SERVICE_INSTALL_PATH"
    fi
    sudo chown root:root "$buildroot/$SERVICE_INSTALL_PATH"
    sudo chmod 644 "$buildroot/$SERVICE_INSTALL_PATH"

    JMX_INSTALL_PATH="$buildroot/opt/opentelemetry-java-contrib-jmx-metrics.jar"
    if [[ -e "$JMX_INSTALL_PATH" ]]; then
        sudo chown root:root "$JMX_INSTALL_PATH"
        sudo chmod 755 "$JMX_INSTALL_PATH"
    fi
}
