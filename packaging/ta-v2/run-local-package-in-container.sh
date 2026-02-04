#!/bin/bash

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

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ASSETS_DIR="${ASSETS_DIR:-${SCRIPT_DIR}/assets}"
LOG_DIR="${SCRIPT_DIR}/local-test-logs/var/log/splunk"
CONTAINER_NAME="${CONTAINER_NAME:-splunk-ta-otel-test}"
SPLUNK_VERSION="${SPLUNK_VERSION:-9.4.0}"

# Check if assets directory exists
if [ ! -d "$ASSETS_DIR" ]; then
    echo "Error: Assets directory not found at $ASSETS_DIR"
    echo "Please run 'make build-otelcol' first to build the collector binaries."
    exit 1
fi

# Stop and remove existing container if it exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Stopping and removing existing container: ${CONTAINER_NAME}"
    docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
fi

# Clean up previous log directory if it exists
if [ -d "$LOG_DIR" ]; then
    echo "Cleaning up previous log directory at $LOG_DIR"
    rm -rf "$LOG_DIR"
fi

# Create log directory
mkdir -p "$LOG_DIR"

# Add the config requirements, if not already present
if [ ! -f "$ASSETS_DIR/local/inputs.conf" ]; then
    mkdir -p "$ASSETS_DIR/local"
    cp "$ASSETS_DIR/default/inputs.conf" "$ASSETS_DIR/local/inputs.conf"
fi

# Check if splunk_access_token is empty and set to test token if needed
if grep -q "^splunk_access_token[[:space:]]*=[[:space:]]*$" "$ASSETS_DIR/local/inputs.conf"; then
    sed -i.bak 's/^splunk_access_token[[:space:]]*=[[:space:]]*$/splunk_access_token = F3K3TestT0Ken/' "$ASSETS_DIR/local/inputs.conf"
    rm -f "$ASSETS_DIR/local/inputs.conf.bak"
fi

echo "Launching Splunk container..."
echo "  Container name: ${CONTAINER_NAME}"
echo "  Splunk version: ${SPLUNK_VERSION}"
echo "  Assets directory: ${ASSETS_DIR}"
echo "  Log directory: ${LOG_DIR}"

# Generate a random password
SPLUNK_PASSWORD="$(uuidgen 2>/dev/null || openssl rand -hex 16)"

# Launch Splunk container
docker run -d --name "${CONTAINER_NAME}" \
    --platform linux/amd64 \
    -e SPLUNK_START_ARGS="--accept-license" \
    -e SPLUNK_PASSWORD="${SPLUNK_PASSWORD}" \
    -v "${ASSETS_DIR}":/opt/splunk/etc/apps/Splunk_TA_OTel_Collector \
    -v "${LOG_DIR}":/opt/splunk/var/log/splunk \
    -p 8000:8000 \
    -p 8088:8088 \
    -p 8089:8089 \
    "splunk/splunk:${SPLUNK_VERSION}"

echo ""
echo "Container launched successfully!"
echo ""

# Wait for splunkd.log to be created
timeout=180
elapsed=0
echo -n "Waiting for splunkd.log creation: "
while [ ! -f "${LOG_DIR}/splunkd.log" ]; do
    if [ "$elapsed" -ge "$timeout" ]; then
        echo "Timeout: splunkd.log was not created within ${timeout} seconds"
        exit 1
    fi
    sleep 2
    elapsed=$((elapsed + 2))
    echo -n "."
done
echo ""

# Wait for Splunk TA OTel Collector to be recorded on the log
timeout=180
elapsed=0
echo -n "Waiting for Splunk_TA_OTel_Collector to be recorded on splunkd.log: "
while ! grep Splunk_TA_OTel_Collector "${LOG_DIR}/splunkd.log" > /dev/null 2>&1; do
    if [ "$elapsed" -ge "$timeout" ]; then
        echo "Timeout: Splunk_TA_OTel_Collector was not recorded on splunkd.log within ${timeout} seconds"
        exit 1
    fi
    sleep 2
    elapsed=$((elapsed + 2))
    echo -n "."
done
echo ""
echo "Splunk_TA_OTel_Collector in splunkd.log:"
grep Splunk_TA_OTel_Collector "${LOG_DIR}/splunkd.log"

echo ""
echo "Splunk Web UI: http://localhost:8000"
echo "  Username: admin"
if [ -t 1 ] || [ "${SPLUNK_PRINT_PASSWORD:-false}" = "true" ]; then
    echo "  Password: ${SPLUNK_PASSWORD}"
else
    echo "  Password: <hidden; set SPLUNK_PRINT_PASSWORD=true to show>"
fi
echo ""
echo "Useful commands:"
echo "  View Splunk logs: docker logs -f ${CONTAINER_NAME}"
echo "  Stop container: docker stop ${CONTAINER_NAME}"
echo "  Remove container: docker rm -f ${CONTAINER_NAME}"
echo "  View splunkd logs: tail -f ${LOG_DIR}/splunkd.log"
echo "  Grep Splunk_TA_OTel_Collector logs: grep Splunk_TA_OTel_Collector \"${LOG_DIR}/splunkd.log\""
echo "  Docker exec shell: docker exec -it -u root ${CONTAINER_NAME} /bin/bash"
echo ""
