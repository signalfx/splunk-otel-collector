#!/bin/bash -eux

set -o pipefail
which jq || (echo "jq not found" && exit 1)
source "${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
BUILD_DIR="$(realpath "$BUILD_DIR")"
TA_FULLPATH="$(repack_with_access_token "$OLLY_ACCESS_TOKEN" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
CI_JOB_ID="${CI_JOB_ID:-$(basename $(dirname "$TA_FULLPATH"))}"
TEST_FOLDER="${TEST_FOLDER:-$BUILD_DIR/$CI_JOB_ID}"
mkdir -p "$TEST_FOLDER"

# Create ORCA container & grab id
splunk_orca -vvv --cloud "${ORCA_CLOUD}" --printer sdd-json --deployment-file "$TEST_FOLDER/orca_deployment.json" --ansible-log "$TEST_FOLDER/ansible-local.log" create --prefix "happypath" --env SPLUNK_CONNECTION_TIMEOUT=600 --platform "$SPLUNK_PLATFORM" --splunk-version "${UF_VERSION}" --local-apps "$TA_FULLPATH" --playbook "$ADDONS_SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM.yml,site.yml"
deployment_id="$(jq -r '.orca_deployment_id' < "$TEST_FOLDER/orca_deployment.json")"
ip_addr="$(jq -r '.server_roles.standalone[0].host' < "$TEST_FOLDER/orca_deployment.json")"

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
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/" "$TEST_FOLDER"
    if safe_grep_log "Starting otel agent" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" &&
       safe_grep_log "Everything is ready" "$TEST_FOLDER/splunk/otel.log"; then
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

# Verify O11y has (recently) received metrics data from this host.  TODO add a resource attribute or similar with the job name
MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
CUTOFF_DELTA='5 min'
export CUTOFF="$(date '+%s%3N' -d "$CUTOFF_DELTA ago")"
otel_hostname="$(grep "host.name" "$TEST_FOLDER/splunk/otel.log" | head -1 | awk -F 'host.name":"' '{print $2}' | awk -F '","' '{print $1}')"
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    curl --header "Content-Type:application/json" --header "X-SF-TOKEN:${OLLY_ACCESS_TOKEN}" "https://api.us0.signalfx.com/v2/metrictimeseries?query=host.name:${otel_hostname}%20AND%20sf_metric:otelcol_process_uptime%20AND%20splunk.distribution:otel-ta" > "$TEST_FOLDER/uptime.json"
    count="$(jq -r '.count' < "$TEST_FOLDER/uptime.json")"
    if [[ "$count" -gt "0" ]] && jq '[.results[].created, .results[].lastUpdated] | max as $max | $max >= ($ENV.CUTOFF | tonumber)' "$TEST_FOLDER/uptime.json" ; then
        break
    fi
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$TEST_FOLDER/splunk/otel.log"
    safe_tail "$TEST_FOLDER/uptime.json"
    echo "Failed to find metrics within $CUTOFF_DELTA after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

# Verify the addon can be restarted successfully
orca_container_name=$(splunk_orca --cloud "${ORCA_CLOUD}" --printer json show --deployment-id "${deployment_id}" containers |  jq -r '.[keys[0]] | .[keys[0]] | .containers | keys[0]')
splunk_orca --cloud "${ORCA_CLOUD}" exec --exec-user splunk "${orca_container_name}" '/opt/splunk/bin/splunk restart'

MAX_ATTEMPTS=12
DELAY=10
ATTEMPT=1
if [ "$PLATFORM"  == "windows" ]; then
    restart_log_file="Splunk_TA_otelutils.log"
else
    restart_log_file="Splunk_TA_otel.log"
fi
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/$restart_log_file" "$TEST_FOLDER/splunk/$restart_log_file"
    # There seems to be an issue on linux where it does not gracefully wait for the job to shut down, need to investigate further.
    (safe_grep_log "INFO Otel agent stop" "$TEST_FOLDER/splunk/$restart_log_file" || safe_grep_log "INFO Stopping otel" "$TEST_FOLDER/splunk/$restart_log_file") && break
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done

if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    cat "$TEST_FOLDER/splunk/$restart_log_file"
    echo "Failed to see restart log after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

# Ensure restart was successful as well
MAX_ATTEMPTS=24
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/Splunk_TA_otel.log" "$TEST_FOLDER/splunk/Splunk_TA_otel.log"
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/otel.log" "$TEST_FOLDER/splunk/otel.log"
    if safe_grep_log "Starting otel agent" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" && safe_grep_log "Everything is ready" "$TEST_FOLDER/splunk/otel.log"; then
        break
    fi
    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    safe_tail "$TEST_FOLDER/splunk/Splunk_TA_otel.log"
    safe_tail "$TEST_FOLDER/splunk/otel.log"
    echo "Failed to see restarted log after $MAX_ATTEMPTS attempts. Logs above."
    exit 1
fi

# Ensure no errors after restart
(grep -qi "ERROR" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$TEST_FOLDER/splunk/otel.log" && exit 1 ) || true

# For release, ensure version is as expected.  TODO move this to another test and compare against tag
actual_version="$(grep "Version" "$TEST_FOLDER/splunk/otel.log" | head -1 | awk -F 'Version": "' '{print $2}' | awk -F '", "' '{print $1}')"
echo "actual version: $actual_version"
EXPECTED_VERSION="v0.128.0"
[[ "$actual_version" != "$EXPECTED_VERSION" ]] && echo "Test failed -- invalid version" && exit 1

# clean up orca container
splunk_orca --cloud ${ORCA_CLOUD} destroy "${deployment_id}"

exit 0
