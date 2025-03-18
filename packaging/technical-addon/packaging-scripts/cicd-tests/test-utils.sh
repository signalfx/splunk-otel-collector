#!/bin/bash -eu

repack_with_access_token() {
    local token="${SPLUNK_O11Y_ACCESS_TOKEN:-$1}"
    local path="${TA_TGZ_PATH:-$2}"

    echo "Adding access token to: $path"

    TEMP_DIR="$BUILD_DIR/repack"
    mkdir -p "$TEMP_DIR"
    TEMP_DIR="$(mktemp -d --tmpdir="$(realpath "$TEMP_DIR")")"
    tar xzvf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" -C "$TEMP_DIR"
    cp -r "$TEMP_DIR/Splunk_TA_otel/default/" "$TEMP_DIR/Splunk_TA_otel/local/"
    chmod -R a+rwx "$TEMP_DIR"
    echo "$token" > "$TEMP_DIR/Splunk_TA_otel/local/access_token"

    random_suffix="$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 6)"
    repacked="$TEMP_DIR/Splunk_TA_otel-${random_suffix}.tgz"
    tar -C "$TEMP_DIR" -hcz --file "$repacked"  "Splunk_TA_otel"
    chmod a+rwx "$repacked"

    echo "$repacked"
    return 0
}

safe_tail() {
    filename="$1"
    set +u
    taillines="$2"
    set -u

    if [ "$taillines" ]; then
        ([ -f "$filename" ] && tail -n "$taillines" "$filename") || echo "File $filename not found"
    else
        ([ -f "$filename" ] && cat "$taillines" "$filename") || echo "File $filename not found"
    fi
}

safe_grep_log() {
    searchstring="$1"
    filename="$2"
    if [ -f "$filename" ]; then
        grep -qi "$searchstring" "$filename"
        return $?
    else
        echo "$filename not found"
        return 1
    fi
}
