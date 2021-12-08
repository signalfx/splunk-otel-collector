#!/usr/bin/env bash

# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#######################################
# Globals
#######################################
CONFDIR="/etc/otel/collector" # Default configuration directory
DIRECTORY= # Either passed as CLI parameter or later set to CONFDIR
TMPDIR="/tmp/splunk-support-bundle-$(date +%s)" # Unique temporary directory for support bundle contents

usage() {
    echo "USAGE: $0 [-h] [-d directory]"
    echo "  -d      directory where Splunk OpenTelemetry Collector configuration is located"
    echo "          (if not specified, defaults to /etc/otel/collector)"
    echo "  -h      display help"
    exit 1
}

#######################################
# Parse command line arguments
#######################################
while [[ $# -gt 0 ]]
do
key="$1"
case $key in
    -d|--directory)
    DIRECTORY="$2"
    shift # past argument
    shift # past value
    ;;
    -t|--tmpdir)
    TMP="$2"
    shift # past argument
    shift # past value
    ;;
    -h|--help)
    usage
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done

#######################################
# Creates a unique temporary directory to store the contents of the support
# bundle. Do not attempt to cleanup to prevent any accidental deletions.
# This command can only be run once per second or will error out.
# This script could result in a lot of temporary data if run multiple times.
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0 if successful, non-zero on error.
#######################################
createTempDir() {
    echo "INFO: Creating temporary directory..."
    # Override primarily for testing
    if [ -n "$TMP" ]; then TMPDIR="$TMP"; fi
    if [ -d "$TMPDIR" ]; then
        echo "ERROR: TMPDIR ($TMPDIR) exists. Exiting."
        exit 1
    else
        mkdir -p "$TMPDIR"
        for d in logs metrics zpages; do
            mkdir -p "$TMPDIR"/$d
        done
    fi
}

#######################################
# Check whether commands exist
# If it doesn't the command output will not be captured
#######################################
checkCommands() {
    echo "INFO: Checking for commands..."
    for EXE in systemctl journalctl curl wget pgrep; do
      if ! command -v $EXE &> /dev/null; then
          echo "WARN: $EXE could not be found."
          echo "      Please install to capture full support bundle."
      fi
    done
}

#######################################
# Gather configuration
# Without this it is very hard to troubleshoot issues so exit if no permissions.
#  - GLOBALS: CONFDIR, DIRECTORY, TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0 if successful, non-zero on error.
#######################################
getConfig() {
    echo "INFO: Getting configuration..."
    # Directory can be passed via CLI parameters
    if [ -z "$DIRECTORY" ]; then
        DIRECTORY=$CONFDIR
    fi
    # If directory does not exist the support bundle is useless so exit
    if [ ! -d "$DIRECTORY" ]; then
        echo "ERROR: Could not find directory ($DIRECTORY)."
        usage
    fi
    # Need to ensure user has permission to access
    if test -r "$DIRECTORY"; then
        cp -r "$DIRECTORY" "$TMPDIR"/config 2>&1
    else
        echo "ERROR: Permission denied to directory ($DIRECTORY)."
        echo "       Run this script with a user who has permissions to this directory."
        exit 1
    fi
    # Also need to get config in memory as dynamic config may modify stored config
    # It's possible user has disabled collecting in memory config
    if timeout 1 bash -c 'cat < /dev/null > /dev/tcp/localhost/55554'; then
        curl -s http://localhost:55554/debug/configz/initial >"$TMPDIR"/config/initial.yaml 2>&1
        curl -s http://localhost:55554/debug/configz/effective >"$TMPDIR"/config/effective.yaml 2>&1
    else
        echo "WARN: localhost:55554 unavailable so in memory configuration not collected"
    fi

}

#######################################
# Gather status
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
getStatus() {
    echo "INFO: Getting status..."
    systemctl status splunk-otel-collector >"$TMPDIR"/logs/splunk-otel-collector.txt 2>&1
    systemctl status td-agent >"$TMPDIR"/logs/td-agent.txt 2>&1
}

#######################################
# Gather logs
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
getLogs() {
    echo "INFO: Getting logs..."
    journalctl -u splunk-otel-collector >"$TMPDIR"/logs/splunk-otel-collector.log 2>&1
    journalctl -u td-agent >"$TMPDIR"/logs/td-agent.log 2>&1
    LOGDIR="/var/log/td-agent"
    if test -r "$LOGDIR"; then
        cp -r /var/log/td-agent "$TMPDIR"/logs/td-agent/ 2>&1
    else
        echo "WARN: Permission denied to directory ($LOGDIR)."
    fi
}

#######################################
# Gather metrics
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
getMetrics() {
    echo "INFO: Getting metric information..."
    # It's possible user has disabled prometheus receiver in metrics pipeline
    if timeout 1 bash -c 'cat < /dev/null > /dev/tcp/localhost/8888'; then
        curl -s http://localhost:8888/metrics >"$TMPDIR"/metrics/collector-metrics.txt 2>&1
    else
        echo "WARN: localhost:8888/metrics unavailable so metrics not collected"
    fi
}

#######################################
# Gather zpages
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
getZpages() {
    echo "INFO: Getting zpages information..."
    # It's possible user has disabled zpages extension
    if timeout 1 bash -c 'cat < /dev/null > /dev/tcp/localhost/55679'; then
        curl -s http://localhost:55679/debug/tracez >"$TMPDIR"/zpages/tracez.html 2>&1
        # Recursively get pages to see output of samples
        wget -q -r -np -l 1 -P "$TMPDIR/zpages" http://localhost:55679/debug/tracez
    else
        echo "WARN: localhost:55679 unavailable so zpages not collected"
    fi
}

#######################################
# Gather Linux information
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0
#######################################
getHostInfo() {
    echo "INFO: Getting host information..."
    # Filter top to only collect Splunk-specific processes
    PIDS=$(pgrep -d "," 'otelcol|fluentd' 2>&1)
    if [ -n "$PIDS" ]; then
        top -b -n 3 -p "$PIDS" >"$TMPDIR"/metrics/top.txt 2>&1
    else
        echo "WARN: Unable to find otelcol or fluentd PIDs"
        echo "      top will not be collected"
    fi
    df -h >"$TMPDIR"/metrics/df.txt 2>&1
    free >"$TMPDIR"/metrics/free.txt 2>&1
}

#######################################
# Tar support bundle
#  - GLOBALS: TMPDIR
#  - ARGUMENTS: None
#  - OUTPUTS: None
#  - RETURN: 0 if successful, non-zero on error
#######################################
tarResults() {
    echo "INFO: Creating tarball..."
    tar cfz "/tmp/$(basename "$TMPDIR").tar.gz" -P "$TMPDIR" 2>&1
    if [ -f "/tmp/$(basename "$TMPDIR").tar.gz" ]; then
        echo "INFO: Support bundle available at: /tmp/$(basename "$TMPDIR").tar.gz"
        echo "      Please attach this to your support case"
        exit 0
    else
        echo "ERROR: Support bundle was not properly created."
        echo "       See $TMPDIR/stdout.log for more information."
        exit 1
    fi
}

main() {
    checkCommands
    getConfig
    getStatus
    getLogs
    getMetrics
    getZpages
    getHostInfo
    tarResults
}

# Attempt to generate a support bundle
# Capture all output
createTempDir
main 2>&1 | tee -a "$TMPDIR"/stdout.log
