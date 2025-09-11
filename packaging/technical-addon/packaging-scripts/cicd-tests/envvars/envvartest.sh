#!/bin/bash -eu
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$ADDONS_SOURCE_DIR" ]] && echo "ADDONS_SOURCE_DIR not set" && exit 1

source "${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
TA_FULLPATH="$(repack_with_test_config "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
rm -rf "$ADDON_DIR/$REPACKED_TA_NAME"

# Set passthrough env vars config & repackage TA
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/passthrough_env_vars.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'gomemlimit=512MiB' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_debug_config_server=test_notused' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_hec_url=test_notused' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_gateway_url=test_notused' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
cp "$ADDONS_SOURCE_DIR/packaging-scripts/cicd-tests/envvars/passthrough_env_vars.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
tar -C "$ADDON_DIR" -hcz --file "$TA_FULLPATH" "Splunk_TA_otel"


echo "Testing with hot TA $TA_FULLPATH ($ADDON_DIR and $REPACKED_TA_NAME)"

DOCKER_COMPOSE_CONFIG="$ADDONS_SOURCE_DIR/packaging-scripts/cicd-tests/envvars/docker-compose.yml"
REPACKED_TA_NAME=$REPACKED_TA_NAME ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" up --quiet-pull --build --force-recreate --wait --detach --timestamps

docker exec -u root envvars-ta-test-envvars-1 /opt/splunk/bin/splunk btool check --debug | grep -qi "Invalid key in stanza" && exit 1

# Most of what we care about can be found in 
# https://github.com/signalfx/splunk-otel-collector/blob/main/internal/settings/settings.go#L40

MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if docker exec -u root envvars-ta-test-envvars-1 grep -qi "Everything is ready. Begin running and processing data." /opt/splunk/var/log/splunk/otel.log; then
        break
    else
        if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
            echo "Failed to see success message in otel.log after $MAX_ATTEMPTS attempts."
            docker exec -u root envvars-ta-test-envvars-1 cat /opt/splunk/var/log/splunk/otel.log
            exit 1
        fi
        echo "success message not found in otel.log Retrying in $DELAY seconds"
        ATTEMPT=$((ATTEMPT + 1))
        sleep $DELAY
    fi
done

echo "checking env vars"
for test_str in \
    'envvartest_given: Str(helloworld)' \
    'envvartest_GOMEMLIMIT: Str(512MiB)' \
    'envvartest_SPLUNK_ACCESS_TOKEN_FILE: Str(/opt/splunk/etc/apps/Splunk_TA_otel/local/access_token)' \
    'envvartest_SPLUNK_API_URL: Str(https://api.us0.signalfx.com)' \
    'envvartest_SPLUNK_BUNDLE_DIR: Str(/opt/splunk/etc/apps/Splunk_TA_otel/linux_x86_64/bin/agent-bundle)' \
    'envvartest_SPLUNK_COLLECTD_DIR: Str(/opt/splunk/etc/apps/Splunk_TA_otel/linux_x86_64/bin/agent-bundle/run/collectd)' \
    'envvartest_SPLUNK_CONFIG: Str(/opt/splunk/etc/apps/Splunk_TA_otel/configs/passthrough_env_vars.yaml)' \
    'envvartest_SPLUNK_DEBUG_CONFIG_SERVER: Str(test_notused)' \
    'envvartest_SPLUNK_GATEWAY_URL: Str(test_notused)' \
    'envvartest_SPLUNK_HEC_URL: Str(test_notused)' \
    'envvartest_SPLUNK_INGEST_URL: Str(https://ingest.us0.signalfx.com)' \
    'envvartest_SPLUNK_LISTEN_INTERFACE: Str(localhost)' \
    'envvartest_SPLUNK_OTEL_LOG_FILE_NAME: Str(/opt/splunk/var/log/splunk/otel.log)' \
    'envvartest_SPLUNK_REALM: Str(us0)' \
; do
    docker exec -u root envvars-ta-test-envvars-1 grep -qi "$test_str" /opt/splunk/var/log/splunk/otel.log
done

REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" down
