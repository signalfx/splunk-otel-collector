#!/bin/bash -eux
set -o pipefail
which jq || (echo "jq not found" && exit 1)
source "${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
BUILD_DIR="$(realpath "$BUILD_DIR")"

CI_JOB_ID="${CI_JOB_ID:-$(mktemp -d)}"
TEST_FOLDER="${TEST_FOLDER:-$BUILD_DIR/$CI_JOB_ID}"
mkdir -p "$TEST_FOLDER"

# customize TA to act as gateway
GATEWAY_TA_FULLPATH="$(repack_with_test_config "$OLLY_ACCESS_TOKEN" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
GATEWAY_REPACKED_TA_NAME="$(basename "$GATEWAY_TA_FULLPATH")"
GATEWAY_ADDON_DIR="$(realpath "$(dirname "$GATEWAY_TA_FULLPATH")")"
rm -rf "$GATEWAY_ADDON_DIR/$GATEWAY_REPACKED_TA_NAME"
# listen on all interfaces (ipv4)
sed -i "s/splunk_listen_interface=localhost/splunk_listen_interface=0.0.0.0/g" "$GATEWAY_ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/ta-gateway-config.yaml' >> "$GATEWAY_ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
tar -C "$GATEWAY_ADDON_DIR" -hcz --file "$GATEWAY_TA_FULLPATH" "Splunk_TA_otel"
echo "Creating splunk cluster with TA $GATEWAY_TA_FULLPATH"

ORCA_SSH_USER="splunk"

gateway_container_info="$(splunk_orca -v --printer json --cloud "$ORCA_CLOUD" --ansible-log "$TEST_FOLDER/ansible-local-gateway.log" create --env SPLUNK_CONNECTION_TIMEOUT=600 --platform "$SPLUNK_PLATFORM" --local-apps "$GATEWAY_TA_FULLPATH" --playbook "$ADDONS_SOURCE_DIR/packaging-scripts/orca-playbook-windows.yml,site.yml")"
echo "$gateway_container_info" > "$TEST_FOLDER/orca-gateway-deployment.json"

# .keys[keys[0]] will grab the first key out of a dict
# Structure is (currently) {"creator":{"deployment":{"containers":{"container_id":{}}}}}
GATEWAY_IPV4_ADDR="$(echo "$gateway_container_info" | jq -r '.[keys[0]] | .[keys[0]] | .containers | .[keys[0]] | .private_address')"

# Customize TA to act as agent which forwards to gateway
GATEWAY_AGENT_TA_FULLPATH="$(repack_with_test_config "$OLLY_ACCESS_TOKEN" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
GATEWAY_AGENT_REPACKED_TA_NAME="$(basename "$GATEWAY_AGENT_TA_FULLPATH")"
GATEWAY_AGENT_ADDON_DIR="$(realpath "$(dirname "$GATEWAY_AGENT_TA_FULLPATH")")"
rm -rf "$GATEWAY_AGENT_ADDON_DIR/$GATEWAY_AGENT_REPACKED_TA_NAME"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/ta-agent-to-gateway-config.yaml' >> "$GATEWAY_AGENT_ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo "splunk_gateway_url=$GATEWAY_IPV4_ADDR" >> "$GATEWAY_AGENT_ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
tar -C "$GATEWAY_AGENT_ADDON_DIR" -hcz --file "$GATEWAY_AGENT_TA_FULLPATH" "Splunk_TA_otel"
echo "Creating splunk cluster with TA $GATEWAY_AGENT_TA_FULLPATH"

gateway_agent_container_info=$(splunk_orca -v --printer json --cloud "$ORCA_CLOUD" --ansible-log "$TEST_FOLDER/ansible-local-gateway-agent.log" create --env SPLUNK_CONNECTION_TIMEOUT=600 --platform "$SPLUNK_PLATFORM"  --local-apps "$GATEWAY_AGENT_TA_FULLPATH" --playbook "$ADDONS_SOURCE_DIR/packaging-scripts/orca-playbook-windows.yml,site.yml")
GATEWAY_AGENT_IPV4_ADDR="$(echo "$gateway_agent_container_info" | jq -r '.[keys[0]] | .[keys[0]] | .containers | .[keys[0]] | .ssh_address')"

echo "$gateway_agent_container_info" > "$TEST_FOLDER/orca-gateway-agent-deployment.json"
GATEWAY_LOGS_DIR="$TEST_FOLDER/$GATEWAY_REPACKED_TA_NAME/"
mkdir -p "$GATEWAY_LOGS_DIR"
GATEWAY_AGENT_LOGS_DIR="$TEST_FOLDER/$GATEWAY_AGENT_REPACKED_TA_NAME/"
mkdir -p "$GATEWAY_AGENT_LOGS_DIR"

# It can take quite some time to extract the agent bundle.  Await for it before trying to pull otel.log.
if [ "$PLATFORM" == "windows" ]; then
    MAX_ATTEMPTS=96
else
    MAX_ATTEMPTS=36
fi
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r "$ORCA_SSH_USER@$GATEWAY_IPV4_ADDR":/opt/splunk/var/log/splunk/ "$GATEWAY_LOGS_DIR"
    if safe_grep_log "Done extracting agent bundle" "$GATEWAY_LOGS_DIR/splunk/Splunk_TA_otel.log"; then
        break
    fi
    echo "Extraction not complete according to Splunk_TA_otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$GATEWAY_LOGS_DIR/splunk/splunkd.log" 200
    safe_tail "$GATEWAY_LOGS_DIR/splunk/Splunk_TA_otel.log"
    echo "Failed to extract agent bundle after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

MAX_ATTEMPTS=12
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r "$ORCA_SSH_USER@$GATEWAY_IPV4_ADDR":/opt/splunk/var/log/splunk/ "$GATEWAY_LOGS_DIR"
    if safe_grep_log "Everything is ready" "$GATEWAY_LOGS_DIR/splunk/otel.log"; then
        break
    fi
    echo "Did not see startup message according to otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$GATEWAY_LOGS_DIR/splunk/Splunk_TA_otel.log"
    safe_tail "$GATEWAY_LOGS_DIR/splunk/otel.log"
    echo "Failed to see startup message in otel.log after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

# It can take quite some time to extract the agent bundle (+7 minutes between start and end log message).  Await for it before trying to pull otel.log.
if [ "$PLATFORM" == "windows" ]; then
    MAX_ATTEMPTS=96
else
    MAX_ATTEMPTS=36
fi
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r "$ORCA_SSH_USER@$GATEWAY_AGENT_IPV4_ADDR":/opt/splunk/var/log/splunk/ "$GATEWAY_AGENT_LOGS_DIR"
    if safe_grep_log "Done extracting agent bundle" "$GATEWAY_AGENT_LOGS_DIR/splunk/Splunk_TA_otel.log"; then
        break
    fi
    echo "Extraction not complete according to Splunk_TA_otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$GATEWAY_AGENT_LOGS_DIR/splunk/splunkd.log" 200
    safe_tail "$GATEWAY_AGENT_LOGS_DIR/splunk/Splunk_TA_otel.log"
    echo "Failed to extract agent bundle after $MAX_ATTEMPTS attempts. Logs above if present."
    exit 1
fi

MAX_ATTEMPTS=24
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r "$ORCA_SSH_USER@$GATEWAY_AGENT_IPV4_ADDR":/opt/splunk/var/log/splunk/ "$GATEWAY_AGENT_LOGS_DIR"
    if safe_grep_log "Everything is ready" "$GATEWAY_AGENT_LOGS_DIR/splunk/otel.log"; then
        break
    fi
    echo "Did not see startup message according to otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    echo "Failed to see startup message in otel.log after $MAX_ATTEMPTS attempts."
    safe_tail "$GATEWAY_AGENT_LOGS_DIR/otel.log"
    exit 1
fi

grep -q "Starting otel agent" "$GATEWAY_LOGS_DIR/splunk/Splunk_TA_otel.log"
(grep -qi "ERROR" "$GATEWAY_LOGS_DIR/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$GATEWAY_LOGS_DIR/splunk/otel.log" && exit 1 ) || true

grep -q "Starting otel agent" "$GATEWAY_AGENT_LOGS_DIR/splunk/Splunk_TA_otel.log"
(grep -qi "ERROR" "$GATEWAY_AGENT_LOGS_DIR/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$GATEWAY_AGENT_LOGS_DIR/splunk/otel.log" && exit 1 ) || true

exit 0
