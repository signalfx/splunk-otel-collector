#!/bin/bash -eux
set -o pipefail
SSH_PORT="22"
if [ "$ORCA_CLOUD" == "kubernetes" ]; then
    SSH_PORT="2222"
fi
which jq || (echo "jq not found" && exit 1)
[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1
BUILD_DIR="$(realpath "$BUILD_DIR")"

source "${SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
TA_FULLPATH="$(repack_with_access_token "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
CI_JOB_ID="${CI_JOB_ID:-$(basename $(dirname "$TA_FULLPATH"))}"
TEST_FOLDER="${TEST_FOLDER:-$BUILD_DIR/$CI_JOB_ID}"
rm -rf "$ADDON_DIR/$REPACKED_TA_NAME"
mkdir -p "$TEST_FOLDER"

# Set discovery specific config & repackage TA
echo 'discovery=true' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'discovery_properties=$SPLUNK_OTEL_TA_HOME/configs/kafkametrics.discovery.properties.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/docker_observer_without_ssl_kafkametrics_config.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
DISCOVERY_SOURCE_DIR="${SOURCE_DIR}/packaging-scripts/cicd-tests/discovery"
cp "$DISCOVERY_SOURCE_DIR/docker_observer_without_ssl_kafkametrics_config.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
cp "$DISCOVERY_SOURCE_DIR/kafkametrics.discovery.properties.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
tar -C "$ADDON_DIR" -hcz --file "$TA_FULLPATH" "Splunk_TA_otel"
echo "Creating splunk cluster with TA $TA_FULLPATH"

ORCA_SSH_USER="splunk"

splunk_orca -vvv --cloud "${ORCA_CLOUD}" --area otel-collector --printer sdd-json --deployment-file "$TEST_FOLDER/orca-deployment.json" --ansible-log "$TEST_FOLDER/ansible-local.log" create --prefix "discovery" --env SPLUNK_CONNECTION_TIMEOUT=600 --platform "$SPLUNK_PLATFORM" --splunk-version "${UF_VERSION}" --local-apps "$TA_FULLPATH" --playbook "$SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM.yml,site.yml" --custom-services "$SOURCE_DIR/packaging-scripts/cicd-tests/agent-bundle/orca-playbook-mysql.yml"

cat "$TEST_FOLDER/orca-deployment.json"

IPV4_ADDR="$(jq -r '.server_roles.standalone[0].host' < "$TEST_FOLDER/orca-deployment.json")"
LOGS_DIR="$TEST_FOLDER/$REPACKED_TA_NAME/"
mkdir -p "$LOGS_DIR"

# It can take quite some time to extract the agent bundle.  Await for it before trying to pull otel.log.
if [ "$PLATFORM" == "windows" ]; then
    MAX_ATTEMPTS=96
else
    MAX_ATTEMPTS=36
fi
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -P "$SSH_PORT" "$ORCA_SSH_USER@$IPV4_ADDR":/opt/splunk/var/log/splunk/ "$LOGS_DIR"
    if safe_grep_log "Done extracting agent bundle" "$LOGS_DIR/splunk/Splunk_TA_otel.log"; then
        break
    fi
    echo "Extraction not complete according to Splunk_TA_otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$LOGS_DIR/splunk/splunkd.log" 200
    safe_tail "$LOGS_DIR/splunk/Splunk_TA_otel.log"
    echo "Failed to extract agent bundle after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

MAX_ATTEMPTS=12
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -P "$SSH_PORT" "$ORCA_SSH_USER@$IPV4_ADDR":/opt/splunk/var/log/splunk/ "$LOGS_DIR"
    if safe_grep_log "Everything is ready" "$LOGS_DIR/splunk/otel.log"; then
        break
    fi
    echo "Did not see startup message according to otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$LOGS_DIR/splunk/Splunk_TA_otel.log"
    safe_tail "$LOGS_DIR/splunk/otel.log"
    echo "Failed to see startup message in otel.log after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

grep -q "Starting otel agent" "$LOGS_DIR/splunk/Splunk_TA_otel.log"
(grep -qi "ERROR" "$LOGS_DIR/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$LOGS_DIR/splunk/otel.log" && exit 1 ) || true


SPLUNK_HOME="/opt/splunk/"

for cmd in \
   "pgrep -f 'discovery'" \
   "pgrep -f 'discovery-properties'" \
   "pgrep -f 'kafkametrics.discovery.properties.yaml'" \
   "test -d $SPLUNK_HOME/etc/apps/Splunk_TA_otel/configs/discovery/config.d.linux" \
; do
    ssh -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P "$SSH_PORT" "$ORCA_SSH_USER@$IPV4_ADDR" "$cmd"
done


MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -i ~/.orca/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -P "$SSH_PORT" "$ORCA_SSH_USER@$IPV4_ADDR":/opt/splunk/var/log/splunk/ "$LOGS_DIR"
    if safe_grep_log "9092" "$LOGS_DIR/splunk/otel.log" && safe_grep_log "kafkametrics receiver is working" "$LOGS_DIR/otel.log" ; then
        break
    fi
    echo "Did not see discovery message according to otel.log... Retrying in $DELAY seconds"
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$LOGS_DIR/splunk/Splunk_TA_otel.log"
    safe_tail "$LOGS_DIR/splunk/otel.log"
    echo "Failed to see discovery message in otel.log after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

exit 0
