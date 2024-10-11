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

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/../common.sh

VERSION="${1:-}"
ARCH="${2:-amd64}"
OUTPUT_DIR="${3:-$REPO_DIR/dist}"
BUNDLE_BASE_DIR="/splunk-otel-collector"
OTELCOL_INSTALL_PATH="$BUNDLE_BASE_DIR/bin/otelcol"

tar_setup_files_and_permissions() {
    local otelcol="$1"
    local config_folder="$2"
    local buildroot="$3"
    local bundle_path="$4"

    create_user_group

    mkdir -p "$buildroot/$(dirname $OTELCOL_INSTALL_PATH)"
    cp -f "$otelcol" "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chown root:root "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$OTELCOL_INSTALL_PATH"

    mkdir -p "$buildroot/$BUNDLE_BASE_DIR/config"
    cp "$config_folder/gateway_config.yaml" "$buildroot/$BUNDLE_BASE_DIR/config/"
    cp "$config_folder/agent_config.yaml" "$buildroot/$BUNDLE_BASE_DIR/config/"

    JMX_METRIC_GATHERER_RELEASE="$(cat $JMX_METRIC_GATHERER_RELEASE_PATH)"
    download_jmx_metric_gatherer "$JMX_METRIC_GATHERER_RELEASE" "$buildroot/$BUNDLE_BASE_DIR"
    JMX_INSTALL_PATH="$buildroot/$BUNDLE_BASE_DIR/opt/opentelemetry-java-contrib-jmx-metrics.jar"
    if [[ -e "$JMX_INSTALL_PATH" ]]; then
        sudo chown root:root "$JMX_INSTALL_PATH"
        sudo chmod 755 "$JMX_INSTALL_PATH"
    fi

    if [[ -n "$bundle_path" ]]; then
        mkdir -p "$buildroot/$BUNDLE_BASE_DIR"
        tar -xzf "$bundle_path" -C "$buildroot/$BUNDLE_BASE_DIR"
        sudo chown -R root:root "$buildroot/$BUNDLE_BASE_DIR"
        sudo chmod -R 755 "$buildroot/$BUNDLE_BASE_DIR"
    fi

}

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi
VERSION="${VERSION#v}"

otelcol_path="$REPO_DIR/bin/otelcol_linux_${ARCH}"
config_folder_path="$REPO_DIR/cmd/otelcol/config/collector"
agent_bundle_path="$REPO_DIR/dist/agent-bundle_linux_${ARCH}.tar.gz"

buildroot="$(mktemp -d)"

tar_setup_files_and_permissions "$otelcol_path" "$config_folder_path" "$buildroot" "$agent_bundle_path"

mkdir -p "$OUTPUT_DIR"

sudo fpm -s dir -t tar -n "${PKG_NAME}_${VERSION}_${ARCH}" -v "$VERSION" -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    "$buildroot/"=/

cd "$OUTPUT_DIR"
gzip -f "${PKG_NAME}_${VERSION}_${ARCH}.tar"
rm -f "${PKG_NAME}_${VERSION}_${ARCH}.tar"
