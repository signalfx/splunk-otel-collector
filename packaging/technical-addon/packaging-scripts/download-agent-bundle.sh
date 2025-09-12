#!/bin/bash -eu
set -o pipefail

[[ -z "$SPLUNK_OTELCOL_DOWNLOAD_BASE" ]] && echo "SPLUNK_OTELCOL_DOWNLOAD_BASE not set" && exit 1

# Download  smart agent
if [ "$PLATFORM" == "windows" ] || [ "$PLATFORM" == "all" ]; then
    SMART_AGENT_BUNDLE="agent-bundle_${OTEL_COLLECTOR_VERSION}_windows_${ARCH}.zip" 
    URL="$SPLUNK_OTELCOL_DOWNLOAD_BASE/v$OTEL_COLLECTOR_VERSION/$SMART_AGENT_BUNDLE"
    mkdir -p "$BUILD_DIR/out/smart-agent/"
    OUTPUT_PATH="$BUILD_DIR/out/smart-agent/$SMART_AGENT_BUNDLE"
    if ! [ -f "$OUTPUT_PATH" ]; then
        wget --no-verbose "$URL" --output-document "$OUTPUT_PATH"
    fi
    echo "SAVED $SMART_AGENT_BUNDLE TO $OUTPUT_PATH"
fi
if [ "$PLATFORM" == "linux" ] || [ "$PLATFORM" == "all" ]; then
    SMART_AGENT_BUNDLE="agent-bundle_${OTEL_COLLECTOR_VERSION}_linux_${ARCH}.tar.gz"
    URL="$SPLUNK_OTELCOL_DOWNLOAD_BASE/v$OTEL_COLLECTOR_VERSION/$SMART_AGENT_BUNDLE"
    mkdir -p "$BUILD_DIR/out/smart-agent/"
    OUTPUT_PATH="$BUILD_DIR/out/smart-agent/$SMART_AGENT_BUNDLE"
    if ! [ -f "$OUTPUT_PATH" ]; then
        wget --no-verbose "$URL" --output-document "$OUTPUT_PATH"
    fi
    echo "SAVED $SMART_AGENT_BUNDLE TO $OUTPUT_PATH"
fi
