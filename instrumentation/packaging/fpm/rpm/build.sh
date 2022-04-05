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
OUTPUT_DIR="${3:-$REPO_DIR/instrumentation/dist}"
LIBSPLUNK_PATH="$REPO_DIR/instrumentation/dist/libsplunk_${ARCH}.so"
JAVA_AGENT_PATH="$REPO_DIR/instrumentation/dist/splunk-otel-javaagent.jar"
JAVA_AGENT_RELEASE="$(cat $JAVA_AGENT_RELEASE_PATH)"

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi

# rpm doesn't like dashes in the version, replace with tildas
VERSION="${VERSION/'-'/'~'}"
VERSION="${VERSION#v}"

download_java_agent "$JAVA_AGENT_RELEASE" "$JAVA_AGENT_PATH"

buildroot="$(mktemp -d)"

setup_files_and_permissions "$LIBSPLUNK_PATH" "$JAVA_AGENT_PATH" "$buildroot"

mkdir -p "$OUTPUT_DIR"

if [[ "$ARCH" = "arm64" ]]; then
    ARCH="aarch64"
elif [[ "$ARCH" = "amd64" ]]; then
    ARCH="x86_64"
fi

sudo fpm -s dir -t rpm -n "$PKG_NAME" -v "$VERSION" -f -p "$OUTPUT_DIR" \
    --vendor "$PKG_VENDOR" \
    --maintainer "$PKG_MAINTAINER" \
    --description "$PKG_DESCRIPTION" \
    --license "$PKG_LICENSE" \
    --url "$PKG_URL" \
    --architecture "$ARCH" \
    --rpm-rpmbuild-define "_build_id_links none" \
    --rpm-summary "$PKG_DESCRIPTION" \
    --rpm-use-file-permissions \
    --after-install "$POSTINSTALL_PATH" \
    --before-remove "$PREUNINSTALL_PATH" \
    --depends sed \
    --depends grep \
    --config-files "$CONFIG_INSTALL_PATH" \
    "$buildroot/"=/

rpm -qpli "${OUTPUT_DIR}/${PKG_NAME}-${VERSION}-1.${ARCH}.rpm"
