#!/bin/bash

# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/../common.sh

VERSION="${1:-}"
ARCH="${2:-amd64}"
OUTPUT_DIR="${3:-$REPO_DIR/dist}"

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi

# rpm doesn't like dashes in the version, replace with underscore
VERSION="${VERSION/'-'/'_'}"
VERSION="${VERSION#v}"

buildroot="$(mktemp -d)"

if [[ "$ARCH" = "arm64" ]]; then
    ARCH="aarch64"
elif [[ "$ARCH" = "amd64" ]]; then
    ARCH="x86_64"
fi

setup_files_and_permissions  "$buildroot"

mkdir -p "$OUTPUT_DIR"

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
    --before-install "$PREINSTALL_PATH" \
    --rpm-posttrans "$POSTINSTALL_PATH" \
    --before-remove "$PREUNINSTALL_PATH" \
    --config-files "$SCRIPT_CONFIG_INSTALL_PATH" \
    --depends python3 \
    --depends gcc-c++ \
    --depends python3-devel \
    "$buildroot/"=/

rpm -qpli "${OUTPUT_DIR}/${PKG_NAME}-${VERSION}*.${ARCH}.rpm"
