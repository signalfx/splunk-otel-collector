#!/bin/bash -eux
set -o pipefail

mkdir -p "$BUILD_DIR"
cp -R "$SOURCE_DIR/Splunk_TA_otel" "$BUILD_DIR"
ls -lAh "$BUILD_DIR"
ls -lAh "$BUILD_DIR/Splunk_TA_otel"
