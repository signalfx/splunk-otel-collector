#!/bin/bash -eux

set -o pipefail
which jq || (echo "jq not found" && exit 1)
source "${SOURCE_DIR}/packaging-scripts/cicd-tests/add-access-token.sh"
BUILD_DIR="$(realpath "$BUILD_DIR")"
TA_FULLPATH="$(repack_with_access_token "$OLLY_ACCESS_TOKEN" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
CI_JOB_ID="${CI_JOB_ID:-$(basename $(dirname "$TA_FULLPATH"))}"
TEST_FOLDER="${TEST_FOLDER:-$BUILD_DIR/$CI_JOB_ID}"
mkdir -p "$TEST_FOLDER"

# Create ORCA container & grab id
splunk_orca -vvv --cloud ${ORCA_CLOUD} --printer sdd-json --deployment-file "$TEST_FOLDER/orca_deployment.json" --ansible-log ansible-local.log create --prefix "happypath" --env SPLUNK_CONNECTION_TIMEOUT=600 --platform $SPLUNK_PLATFORM --splunk-version "${UF_VERSION}" --local-apps "$TA_FULLPATH" --playbook "$SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM.yml,site.yml"
# TODO use jq not awk
deployment_id="$(grep "orca_deployment_id" "$TEST_FOLDER/orca_deployment.json" | awk -F ':' '{print $2}' | awk -F '"' '{print $2}')"
echo "$deployment_id" > "$TEST_FOLDER/deployment_id.txt"
ip_addr="$(jq -r '.server_roles.standalone[0].host' < "$TEST_FOLDER/orca_deployment.json")"

if [ "$PLATFORM" == "windows" ]; then
    # Windows takes forever to extract
    echo "sleeping for 700s at $(date)"
    sleep 700s
else
    # Can likely drop this way down, but give the collector time to collect metrics/traces
    echo "sleeping for 90s at $(date)"
    sleep 90s
fi


# Copy logs from container
scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/" "$TEST_FOLDER"

# Verify Otel agent is running without any error
grep -q "Starting otel agent" "$TEST_FOLDER/splunk/Splunk_TA_otel.log"
grep -q "Everything is ready" "$TEST_FOLDER/splunk/otel.log"
(grep -qi "ERROR" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$TEST_FOLDER/splunk/otel.log" && exit 1 ) || true

# Verify Olly has received metrics data from this host
MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
export CUTOFF="$(date '+%s%3N' -d '5 min ago')"
otel_hostname="$(grep "host.name" "$TEST_FOLDER/splunk/otel.log" | head -1 | awk -F 'host.name":"' '{print $2}' | awk -F '","' '{print $1}')"
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    curl --header "Content-Type:application/json" --header "X-SF-TOKEN:${OLLY_ACCESS_TOKEN}" "https://api.us0.signalfx.com/v2/metrictimeseries?query=host.name:${otel_hostname}%20AND%20sf_metric:otelcol_process_uptime%20AND%20splunk.distribution:otel-ta" > "$TEST_FOLDER/uptime.json"
    count=$( grep '"count"' "$TEST_FOLDER/uptime.json" | awk -F ':\ ' '{print $2}' | awk -F ',' '{print $1}')

    if [[ "$count" -gt "0" ]] && jq '[.results[].created, .results[].lastUpdated] | max as $max | $max >= ($ENV.CUTOFF | tonumber)' "$TEST_FOLDER/uptime.json" ; then
        break
    else
        ATTEMPT=$((ATTEMPT + 1))
        sleep $DELAY
    fi
done
if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    echo "Failed find metrics in last 5m after $MAX_ATTEMPTS attempts."
    cat "$GATEWAY_AGENT_LOGS_DIR/otel.log"
    exit 1
fi

# Verify the addon can be restarted successfully
orca_container_name=$(splunk_orca --cloud ${ORCA_CLOUD} --printer json show --deployment-id "${deployment_id}" containers |  jq -r '.[keys[0]] | .[keys[0]] | .containers | keys[0]')
splunk_orca --cloud "${ORCA_CLOUD}" exec --exec-user splunk "${orca_container_name}" '/opt/splunk/bin/splunk restart'
sleep 90s

MAX_ATTEMPTS=30
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/Splunk_TA_otel.log" "$TEST_FOLDER/"
    if [ "$PLATFORM"  == "windows" ]; then
        scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/Splunk_TA_otelutils.log" "$TEST_FOLDER/"
        grep -q "INFO Otel agent stop" "$TEST_FOLDER/Splunk_TA_otelutils.log" && break
    else
        grep -q "INFO Otel agent stop" "$TEST_FOLDER/Splunk_TA_otel.log" && break
    fi

    ATTEMPT=$((ATTEMPT + 1))
    sleep $DELAY
done

if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
    echo "Failed to see restart log m after $MAX_ATTEMPTS attempts."
    if [ "$PLATFORM"  == "windows" ]; then
        cat "$TEST_FOLDER/Splunk_TA_otelutils.log"
    else
        cat "$TEST_FOLDER/Splunk_TA_otel.log"
    fi
    exit 1
fi

scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r -i ~/.orca/id_rsa "splunk@$ip_addr:/opt/splunk/var/log/splunk/otel.log" "$TEST_FOLDER/"
grep -q "Starting otel agent" "$TEST_FOLDER/splunk/Splunk_TA_otel.log"
grep -q "Everything is ready" "$TEST_FOLDER/splunk/otel.log"
(grep -qi "ERROR" "$TEST_FOLDER/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$TEST_FOLDER/splunk/otel.log" && exit 1 ) || true

# Ensure version is as expected
actual_version="$(grep "Version" "$TEST_FOLDER/splunk/otel.log" | head -1 | awk -F 'Version": "' '{print $2}' | awk -F '", "' '{print $1}')"
echo "actual version: $actual_version"
[[ "$actual_version" != "v0.111.0" ]] && echo "Test failed -- invalid version" && exit 1

# clean up orca container
splunk_orca --cloud ${ORCA_CLOUD} destroy "${deployment_id}"

exit 0
