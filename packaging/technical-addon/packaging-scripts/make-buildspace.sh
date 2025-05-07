#!/bin/bash -eux
set -o pipefail

mkdir -p "$BUILD_DIR"
cp -R "$ADDONS_SOURCE_DIR/Splunk_TA_otel" "$BUILD_DIR"
mkdir -p "$BUILD_DIR/configs/discovery"
cp -R "$ADDONS_SOURCE_DIR/../../cmd/otelcol/config/collector/config.d.linux" "$BUILD_DIR/configs/discovery"
