#!/bin/bash -eu

repack_with_test_config() {
    local token="${SPLUNK_O11Y_ACCESS_TOKEN:-$1}"
    local path="${TA_TGZ_PATH:-$2}"

    echo "Setting test config to: $path"

    TEMP_DIR="$BUILD_DIR/repack"
    mkdir -p "$TEMP_DIR"
    TEMP_DIR="$(mktemp -d --tmpdir="$(realpath "$TEMP_DIR")")"
    tar xzvf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" -C "$TEMP_DIR"
    cp -r "$TEMP_DIR/Splunk_TA_otel/default/" "$TEMP_DIR/Splunk_TA_otel/local/"
    chmod -R a+rwx "$TEMP_DIR"

    echo "$token" > "$TEMP_DIR/Splunk_TA_otel/local/access_token"

    # Loop over all YAML files and update log level and output lines
    for yaml_file in "$TEMP_DIR/Splunk_TA_otel/local/"*.yaml; do
        if [ -f "$yaml_file" ]; then
            sed -i "s/^level: .*/level: debug/" "$yaml_file"
            sed -i "s/^# output_paths: /output_paths: /" "$yaml_file"
        fi
    done

    random_suffix="$(LC_CTYPE=c tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 6)"
    repacked="$TEMP_DIR/Splunk_TA_otel-${random_suffix}.tgz"
    COPYFILE_DISABLE=1 tar --format ustar -C "$TEMP_DIR" -hcz --file "$repacked"  "Splunk_TA_otel"
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
        ([ -f "$filename" ] && cat "$filename") || echo "File $filename not found"
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
