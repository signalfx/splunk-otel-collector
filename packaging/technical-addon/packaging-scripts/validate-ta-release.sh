#!/bin/bash -eux

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR MUST BE SET" && exit 1
[[ -z "$ADDONS_SOURCE_DIR" ]] && echo "ADDONS_SOURCE_DIR MUST BE SET" && exit 1
[[ -z "$OTEL_COLLECTOR_VERSION" ]] && echo "OTEL_COLLECTOR_VERSION MUST BE SET" && exit 1


TA_VERSION=$(sed -n '/^\[launcher\]/,/^\[/{s/^version = \(.*\)/\1/p}' "$ADDONS_SOURCE_DIR/Splunk_TA_otel/default/app.conf" | tr -d ' \t\n\r')
[[ -z "$TA_VERSION" ]] && echo "TA_VERSION MUST BE SET" && exit 1

TEMPDIR="$(mktemp -d)"
cp "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" "$BUILD_DIR/out/distribution/Splunk_TA_otel-${TA_VERSION}.tgz"


cp "$BUILD_DIR/out/distribution/Splunk_TA_otel-${TA_VERSION}.tgz" "$TEMPDIR"
echo "inspecting addon in $TEMPDIR"
cd "$TEMPDIR"
tar -xzvf "Splunk_TA_otel-${TA_VERSION}.tgz"

access_token_size="$(ls --size "Splunk_TA_otel/default/access_token")"
[[ "$access_token_size" != "0 Splunk_TA_otel/default/access_token" ]] && echo "access token is not empty! validation failed" && exit 1

[ ! -f ./Splunk_TA_otel/windows_x86_64/bin/otelcol_windows_amd64.exe ] && echo "Can't find windows binary in Addon" && exit 1
ACTUAL_VERSION="$(Splunk_TA_otel/linux_x86_64/bin/otelcol_linux_amd64 --version)"
EXPECTED_VERSION="otelcol version v$OTEL_COLLECTOR_VERSION"
[[ "$EXPECTED_VERSION" != "$ACTUAL_VERSION" ]] && echo "Invalid version -- Expected $EXPECTED_VERSION but got $ACTUAL_VERSION" && exit 1

echo "Validation passed for $ACTUAL_VERSION"