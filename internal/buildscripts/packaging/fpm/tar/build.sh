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
SMART_AGENT_RELEASE="${4:-}"
BUNDLE_BASE_DIR="/splunk-otel-collector"
AGENT_BUNDLE_INSTALL_DIR="$BUNDLE_BASE_DIR/agent-bundle"
OTELCOL_INSTALL_PATH="$BUNDLE_BASE_DIR/bin/otelcol"
TRANSLATESFX_INSTALL_PATH="$BUNDLE_BASE_DIR/bin/translatesfx"

tar_download_smart_agent() {
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
    tar -xzf "$buildroot/signalfx-agent.tar.gz" -C "$buildroot/"
    mv "$buildroot/signalfx-agent" "$buildroot/$AGENT_BUNDLE_INSTALL_DIR"
    find "$buildroot/$AGENT_BUNDLE_INSTALL_DIR" -wholename "*test*.key" -delete -or -wholename "*test*.pem" -delete
    rm -f "$buildroot/signalfx-agent.tar.gz"
}

tar_setup_files_and_permissions() {
    local otelcol="$1"
    local translatesfx="$2"
    local config_folder="$3"
    local buildroot="$4"

    create_user_group

    mkdir -p "$buildroot/$(dirname $OTELCOL_INSTALL_PATH)"
    cp -f "$otelcol" "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chown root:root "$buildroot/$OTELCOL_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$OTELCOL_INSTALL_PATH"

    mkdir -p "$buildroot/$BUNDLE_BASE_DIR/config"
    cp "$config_folder/gateway_config.yaml" "$buildroot/$BUNDLE_BASE_DIR/config/"
    cp "$config_folder/agent_config.yaml" "$buildroot/$BUNDLE_BASE_DIR/config/"

    mkdir -p "$buildroot/$(dirname $TRANSLATESFX_INSTALL_PATH)"
    cp -f "$translatesfx" "$buildroot/$TRANSLATESFX_INSTALL_PATH"
    sudo chown root:root "$buildroot/$TRANSLATESFX_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$TRANSLATESFX_INSTALL_PATH"
}

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi
VERSION="${VERSION#v}"

if [[ -z "$SMART_AGENT_RELEASE" ]]; then
    SMART_AGENT_RELEASE="$(cat $SMART_AGENT_RELEASE_PATH)"
fi

otelcol_path="$REPO_DIR/bin/otelcol_linux_${ARCH}"
translatesfx_path="$REPO_DIR/bin/translatesfx_linux_${ARCH}"
config_folder_path="$REPO_DIR/cmd/otelcol/config/collector"

buildroot="$(mktemp -d)"

if [ "$ARCH" = "amd64" ]; then
    tar_download_smart_agent "$SMART_AGENT_RELEASE" "$buildroot"
fi

tar_setup_files_and_permissions "$otelcol_path" "$translatesfx_path" "$config_folder_path" "$buildroot"

mkdir -p "$OUTPUT_DIR"

sudo fpm -s dir -t tar -n "${PKG_NAME}_${VERSION}_${ARCH}" -v "$VERSION" -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    "$buildroot/"=/

cd "$OUTPUT_DIR"
gzip "${PKG_NAME}_${VERSION}_${ARCH}.tar"