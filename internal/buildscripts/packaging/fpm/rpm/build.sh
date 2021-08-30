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

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi

# rpm doesn't like dashes in the version, replace with tildas
VERSION="${VERSION/'-'/'~'}"
VERSION="${VERSION#v}"

if [[ -z "$SMART_AGENT_RELEASE" ]]; then
    SMART_AGENT_RELEASE="$(cat $SMART_AGENT_RELEASE_PATH)"
fi

otelcol_path="$REPO_DIR/bin/otelcol_linux_${ARCH}"

buildroot="$(mktemp -d)"

if [[ "$ARCH" = "arm64" ]]; then
    ARCH="aarch64"
elif [[ "$ARCH" = "amd64" ]]; then
    ARCH="x86_64"
    download_smart_agent "$SMART_AGENT_RELEASE" "$buildroot"
fi

setup_files_and_permissions "$otelcol_path" "$buildroot"

mkdir -p "$OUTPUT_DIR"

sudo fpm -s dir -t rpm -n "$PKG_NAME" -v "$VERSION" -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    --architecture "$ARCH" \
    --rpm-summary "$PKG_DESCRIPTION" \
    --rpm-use-file-permissions \
    --before-install "$PREINSTALL_PATH" \
    --after-install "$POSTINSTALL_PATH" \
    --before-remove "$PREUNINSTALL_PATH" \
    --config-files "$AGENT_CONFIG_INSTALL_PATH" \
    --config-files "$GATEWAY_CONFIG_INSTALL_PATH" \
    --config-files "$FLUENTD_CONFIG_INSTALL_DIR" \
    "$buildroot/"=/

rpm -qpli "${OUTPUT_DIR}/${PKG_NAME}-${VERSION}*.${ARCH}.rpm"
