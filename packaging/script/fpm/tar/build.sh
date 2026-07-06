#!/bin/bash

# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/../common.sh

VERSION="${1:-}"
ARCH="${2:-amd64}"
OUTPUT_DIR="${3:-$REPO_DIR/dist}"
BUNDLE_BASE_DIR="/splunk-otel-script"

tar_setup_files_and_permissions() {
    local buildroot="$1"

    create_user_group

    mkdir -p "$buildroot/$BUNDLE_BASE_DIR"

    cp "$REPO_DIR/cron.py" "$buildroot/$BUNDLE_BASE_DIR"
    cp "$REPO_DIR/requirements.txt" "$buildroot/$BUNDLE_BASE_DIR"

    sudo chown $SERVICE_USER:$SERVICE_GROUP "$buildroot/$BUNDLE_BASE_DIR"

    cp -r "$FPM_DIR/etc" "$buildroot/etc"
    sudo chown -R $SERVICE_USER:$SERVICE_GROUP "$buildroot/etc/otel"
    sudo chmod -R 755 "$buildroot/etc/otel"
    sudo chmod 600 "$buildroot/etc/otel/script/$SERVICE_NAME.conf"

}

if [[ -z "$VERSION" ]]; then
    VERSION="$( get_version )"
fi
VERSION="${VERSION#v}"

buildroot="$(mktemp -d)"

tar_setup_files_and_permissions "$buildroot"

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
