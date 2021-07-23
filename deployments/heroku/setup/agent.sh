#!/bin/bash

if [ "$DYNOTYPE" == "run" ]; then
    exit 0
fi

# Set configuration file

export SPLUNK_CONFIG_DIR="$HOME/.splunk"
export SPLUNK_COLLECTD_CONFIG_DIR="$SPLUNK_CONFIG_DIR/signalfx-agent/var/run/collectd"
mkdir -p "$SPLUNK_COLLECTD_CONFIG_DIR"

export FALLBACK_AGENT_CONFIG="$SPLUNK_CONFIG_DIR/config.yaml"
export DEFAULT_APP_CONFIG="$HOME/config.yaml"

if [[ -f "$DEFAULT_APP_CONFIG" ]]; then
    export SPLUNK_CONFIG="${SPLUNK_CONFIG-$DEFAULT_APP_CONFIG}"
else
    # Can be overridden by an envvar
    export SPLUNK_CONFIG="${SPLUNK_CONFIG-$FALLBACK_AGENT_CONFIG}"
fi

# Set other env vars

if [[ -z "$SPLUNK_API_URL" ]]; then
    export SPLUNK_API_URL="https://api.$SPLUNK_REALM.signalfx.com"
fi
if [[ -z "$SPLUNK_INGEST_URL" ]]; then
    export SPLUNK_INGEST_URL="https://ingest.$SPLUNK_REALM.signalfx.com"
fi
if [[ -z "$SPLUNK_TRACE_URL" ]]; then
    export SPLUNK_TRACE_URL="https://ingest.$SPLUNK_REALM.signalfx.com/v2/trace"
fi

export SPLUNK_BUNDLE_DIR="$SPLUNK_CONFIG_DIR/signalfx-agent"

if [[ -z "$SPLUNK_LOG_FILE" ]]; then
    export SPLUNK_LOG_FILE=/dev/stdout
else
    mkdir -p $(dirname $SPLUNK_LOG_FILE)
fi

# Start connector

chmod a+x $SPLUNK_CONFIG_DIR/otelcol_linux_amd64
$SPLUNK_CONFIG_DIR/otelcol_linux_amd64 > $SPLUNK_LOG_FILE 2>&1&
