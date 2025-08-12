#!/bin/bash -eux
set -o pipefail
SSH_PORT="22"
if [ "$ORCA_CLOUD" == "kubernetes" ]; then
    SSH_PORT="2222"
fi
which jq || (echo "jq not found" && exit 1)
source "${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
BUILD_DIR="$(realpath "$BUILD_DIR")"
TA_FULLPATH="$(repack_with_access_token "$OLLY_ACCESS_TOKEN" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
CI_JOB_ID="${CI_JOB_ID:-$(basename $(dirname "$TA_FULLPATH"))}"
TEST_FOLDER="${TEST_FOLDER:-$BUILD_DIR/$CI_JOB_ID}"
mkdir -p "$TEST_FOLDER"

# Customize TA to use mysql collectd reciever
#rm -rf "$ADDON_DIR/$REPACKED_TA_NAME"
rm -rf "$TA_FULLPATH"
cp "$ADDONS_SOURCE_DIR/packaging-scripts/cicd-tests/agent-bundle/ta-agent-mysql-config.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/ta-agent-mysql-config.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
tar -C "$ADDON_DIR" -hcz --file "$TA_FULLPATH" "Splunk_TA_otel"
echo "Creating orca cluster with nginx and TA $TA_FULLPATH"

# Create ORCA container & grab id
splunk_orca -vvv --cloud "${ORCA_CLOUD}" --area otel-collector --printer sdd-json --deployment-file "$TEST_FOLDER/orca_deployment.json" --ansible-log "$TEST_FOLDER/ansible-local.log" create --prefix "collectd" --env SPLUNK_CONNECTION_TIMEOUT=600 --platform "$SPLUNK_PLATFORM" --splunk-version "${UF_VERSION}" --local-apps "$TA_FULLPATH" --playbook "$ADDONS_SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM.yml,site.yml"
cat "$TEST_FOLDER/orca_deployment.json"
deployment_id="$(jq -r '.orca_deployment_id' < "$TEST_FOLDER/orca_deployment.json")"
ip_addr="$(jq -r '.server_roles.standalone[0].host' < "$TEST_FOLDER/orca_deployment.json")"

# Seed in the correct IP address for the TA
SPLUNK_HOME="/opt/splunk" # as it's ansible ssh, path is linux-like with c==root
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ~/.orca/id_rsa "ssh://splunk@$ip_addr:$SSH_PORT" '/opt/splunk/bin/splunk restart'


# Check for successful startup
if [ "$PLATFORM" == "windows" ]; then
    MAX_ATTEMPTS=96 # Windows takes a long time to extract, often 7 minutes on default hardware
else
    MAX_ATTEMPTS=36
fi
ATTEMPT=1
DELAY=10
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    # Copy logs from container
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ~/.orca/id_rsa -r -P $SSH_PORT "splunk@$ip_addr:/opt/splunk/var/log/splunk/" "$TEST_FOLDER"
    if safe_grep_log "Starting otel agent" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" && \
        safe_grep_log "Everything is ready" "$TEST_FOLDER/splunk/otel.log" ; then
        break
    fi
    ATTEMPT=$((ATTEMPT + 1))
    sleep "$DELAY"
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$TEST_FOLDER/splunk/splunkd.log" 200
    safe_tail "$TEST_FOLDER/splunk/Splunk_TA_otel.log"
    safe_tail "$TEST_FOLDER/splunk/otel.log"
    echo "Failed to find successful startup message(s) after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

# Verify Otel agent is running without any error
(grep -qi "ERROR" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$TEST_FOLDER/splunk/otel.log" && exit 1 ) || true

exit 0

