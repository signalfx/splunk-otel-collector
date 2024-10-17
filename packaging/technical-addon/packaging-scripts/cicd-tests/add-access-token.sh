#!/bin/bash -eu

repack_with_access_token() {
    local token="${SPLUNK_O11Y_ACCESS_TOKEN:-$1}"
    local path="${TA_TGZ_PATH:-$2}"

    echo "Adding access token to: $path"

    TEMP_DIR="$BUILD_DIR/repack"
    mkdir -p "$TEMP_DIR"
    tar xzvf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" -C "$TEMP_DIR"
    cp -r "$TEMP_DIR/Splunk_TA_otel/default/" "$TEMP_DIR/Splunk_TA_otel/local/"
    echo "$token" > "$TEMP_DIR/Splunk_TA_otel/local/access_token"

    random_suffix="$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 6)"
    repacked="$TEMP_DIR/Splunk_TA_otel-${random_suffix}.tgz"
    tar -C "$TEMP_DIR" -hcz --file "$repacked"  "Splunk_TA_otel"

    echo "$repacked"
    return 0
}