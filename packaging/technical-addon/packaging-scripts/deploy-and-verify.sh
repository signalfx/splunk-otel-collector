#!/bin/bash -eux

set -o pipefail

# Grab artifact from the build stage, add token and run the tests in orca container
cd "$BUILD_DIR"
tar xzf "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz"
cd "Splunk_TA_otel"
cp -r default/ local/
echo "$OLLY_ACCESS_TOKEN" >> "local/access_token"

# Re-package it & clean up
cd "-"
mkdir -p "$BUILD_DIR/ci-cd"
if [ "$PLATFORM" == "windows" ]; then
  rm -Rf "./Splunk_TA_otel/linux_x86_64"
elif [ "$PLATFORM" == "linux" ]; then
  rm -Rf "./Splunk_TA_otel/windows_x86_64"
fi
tar -czvf "$BUILD_DIR/ci-cd/Splunk_TA_otel.tgz" "./Splunk_TA_otel"

# Create ORCA container & grab id
splunk_orca -vvv --cloud ${ORCA_CLOUD} --printer sdd-json --deployment-file "$BUILD_DIR/orca_deployment.json" --ansible-log ansible-local.log create --env SPLUNK_CONNECTION_TIMEOUT=600 --so 1 --splunk-version "${UF_VERSION}" --platform "${SPLUNK_PLATFORM}" --local-apps "$BUILD_DIR/ci-cd/Splunk_TA_otel.tgz" --cloud-instance-type-mapping '{"universal_forwarder": "c5.large", "heavy_forwarder": "c5.large", "indexer": "c5.large", "eventgen_standalone": "c5.large", "testrunner": "c5.large", "license_master": "c5.large", "deployment_server": "c5.large", "deployer": "c5.large", "eventgen_server": "c5.large", "cluster_master": "c5.large", "search_head": "c5.large", "eventgen_controller": "c5.large", "standalone": "c5.large"}' --playbook "$SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM.yml,site.yml"
deployment_id="$(grep "orca_deployment_id" "$BUILD_DIR/orca_deployment.json" | awk -F ':' '{print $2}' | awk -F '"' '{print $2}')"
echo "deployment_id: $deployment_id"
splunk_orca --cloud ${ORCA_CLOUD} show containers --deployment-id "${deployment_id}" >> "$BUILD_DIR/container_details.txt"
orca_container=$(grep "Container Name" "$BUILD_DIR/container_details.txt" | awk -F " " '{print $8}')

if [ "$PLATFORM" == "windows" ]; then
    # Windows takes forever to extract
    sleep 700s
else
    # Can likely drop this way down, but give the collector time to collect metrics/traces
    sleep 90s
fi


# Copy logs from container
mkdir -p "$BUILD_DIR"/splunklog
splunk_orca --cloud "${ORCA_CLOUD}" copy from "${orca_container}" ${ORCA_OPTION} --source /opt/splunk/var/log/splunk/ --destination "$BUILD_DIR/splunklog"
set +u
if [ -n "$SPLUNK_OTEL_TA_DEBUG" ]; then
    set -u
    mkdir "$BUILD_DIR/otel-ta-deployment"
    splunk_orca --cloud "${ORCA_CLOUD}" copy from "${orca_container}" ${ORCA_OPTION} --source /opt/splunk/etc/apps/Splunk_TA_otel/ --destination "$BUILD_DIR/otel-ta-deployment/Splunk_TA_otel"
fi
set -u

# Verify Otel agent is running without any error
grep -q "Starting otel agent" "$BUILD_DIR/splunklog/splunk/Splunk_TA_otel.log"
grep -q "Everything is ready" "$BUILD_DIR/splunklog/splunk/otel.log"
(grep -qi "ERROR" "$BUILD_DIR/splunklog/splunk/Splunk_TA_otel.log" && exit 1 ) || true
(grep -qi "ERROR" "$BUILD_DIR/splunklog/splunk/otel.log" && exit 1 ) || true

# Verify Olly has received metrics data from this host
otel_hostname="$(grep "host.name" "$BUILD_DIR/splunklog/splunk/otel.log" | head -1 | awk -F 'host.name":"' '{print $2}' | awk -F '","' '{print $1}')"
echo "otel hostname: $otel_hostname"
#curl --header "Content-Type:application/json" --header "X-SF-TOKEN:${OLLY_ACCESS_TOKEN}" "https://api.us0.signalfx.com/v2/metrictimeseries?query=host.name:${otel_hostname}%20AND%20sf_metric:cpu_utilization%20AND%20sf_lastActiveMs:%3E%3D$(date '+%s%3N' -d '5 min ago')" > cpu_util.json
curl --header "Content-Type:application/json" --header "X-SF-TOKEN:${OLLY_ACCESS_TOKEN}" "https://api.us0.signalfx.com/v2/metrictimeseries?query=host.name:${otel_hostname}%20AND%20sf_metric:cpu.utilization" > cpu_util.json
count=$(grep '"count"' cpu_util.json | awk -F ':\ ' '{print $2}' | awk -F ',' '{print $1}')
echo $count
[[ "$count" -eq "0" ]] && echo "Test failed -- could not find cpu utilization metrics" && exit 1
CUTOFF="$(date '+%s%3N' -d '5 min ago')"
export CUTOFF
jq '[.results[].created, .results[].lastUpdated] | max as $max | $max >= ($ENV.CUTOFF | tonumber)' cpu_util.json

# Verify the current otel process was killed by PID by restarting splunk
if [ "$PLATFORM" == "windows" ]; then
  splunk_orca --cloud "${ORCA_CLOUD}" provision "${deployment_id}" --playbooks "$SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM-restart.yml"
  splunk_orca --cloud "${ORCA_CLOUD}" copy from "${orca_container}" ${ORCA_OPTION} --source /opt/splunk/var/log/splunk/Splunk_TA_otelutils.log --destination "$BUILD_DIR/splunklog"
  grep -q "INFO Otel agent stopped" "$BUILD_DIR/splunklog/Splunk_TA_otelutils.log" || exit 1
elif [ "$PLATFORM" == "linux" ]; then
  splunk_orca --cloud "${ORCA_CLOUD}" provision "${deployment_id}" --playbooks "$SOURCE_DIR/packaging-scripts/orca-playbook-$PLATFORM-restart.yml"
  splunk_orca --cloud "${ORCA_CLOUD}" copy from "${orca_container}" ${ORCA_OPTION} --source /opt/splunk/var/log/splunk/Splunk_TA_otel.log --destination "$BUILD_DIR/splunklog"
  grep -q "Otel agent stopped" "$BUILD_DIR/splunklog/Splunk_TA_otel.log" || exit 1
fi

# clean up orca container
splunk_orca --cloud ${ORCA_CLOUD} destroy "${deployment_id}"

# Ensure telemetry is coming along
curl --header "Content-Type:application/json" --header "X-SF-TOKEN:${OLLY_ACCESS_TOKEN}" "https://api.us0.signalfx.com/v2/metrictimeseries?query=host.name:${otel_hostname}%20AND%20sf_metric:otelcol_process_uptime%20AND%20splunk.distribution:otel-ta" > uptime.json
count=$( grep '"count"' uptime.json | awk -F ':\ ' '{print $2}' | awk -F ',' '{print $1}')
echo $count
[[ "$count" -eq "0" ]] && echo "Test failed -- could not find uptime metrics" && exit 1
jq '[.results[].created, .results[].lastUpdated] | max as $max | $max >= ($ENV.CUTOFF | tonumber)' uptime.json

# Ensure version is as expected
actual_version="$(grep "Version" "$BUILD_DIR/splunklog/splunk/otel.log" | head -1 | awk -F 'Version": "' '{print $2}' | awk -F '", "' '{print $1}')"
echo "actual version: $actual_version"
[[ "$actual_version" != "v0.104.0" ]] && echo "Test failed -- invalid version" && exit 1

# Ensure gateway mode (and any other tests we wish to run) works
"$SOURCE_DIR/packaging-scripts/cicd-tests/gateway.sh" || (echo "failed to verify gateway" && exit 1)

if [ "$PLATFORM" == "linux" ]; then
  "$SOURCE_DIR/packaging-scripts/cicd-tests/collectd.sh" || (echo "failed to verify collectd" && exit 1)
fi

exit 0
