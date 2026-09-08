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
OTELCOLLAUNCHER_INSTALL_PATH="$BUNDLE_BASE_DIR/bin/otelcollauncher"
OPAMPSUPERVISOR_INSTALL_PATH="$BUNDLE_BASE_DIR/bin/opampsupervisor"

tar_setup_files_and_permissions() {
    local otelcol="$1"
    local otelcollauncher="$2"
    local opampsupervisor="$3"
    local config_folder="$4"
    local buildroot="$5"

    create_user_group

    install_binary "$otelcol" "$OTELCOL_INSTALL_PATH" "$buildroot"
    install_binary "$otelcollauncher" "$OTELCOLLAUNCHER_INSTALL_PATH" "$buildroot"
    install_binary "$opampsupervisor" "$OPAMPSUPERVISOR_INSTALL_PATH" "$buildroot"

    mkdir -p "$buildroot/$BUNDLE_BASE_DIR/config"
    cp "$config_folder/gateway_config.yaml" "$buildroot/$BUNDLE_BASE_DIR/config/"
    cp "$config_folder/agent_config.yaml" "$buildroot/$BUNDLE_BASE_DIR/config/"
}

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi
VERSION="${VERSION#v}"

otelcol_path="$REPO_DIR/bin/otelcol_linux_${ARCH}"
otelcollauncher_path="$REPO_DIR/bin/otelcollauncher_linux_${ARCH}"
opampsupervisor_path="$REPO_DIR/bin/opampsupervisor_linux_${ARCH}"
config_folder_path="$REPO_DIR/cmd/otelcol/config/collector"

buildroot="$(mktemp -d)"

tar_setup_files_and_permissions \
    "$otelcol_path" \
    "$otelcollauncher_path" \
    "$opampsupervisor_path" \
    "$config_folder_path" \
    "$buildroot"

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
