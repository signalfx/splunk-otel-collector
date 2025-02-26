#!/bin/bash -eu
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1

source "${SOURCE_DIR}/packaging-scripts/cicd-tests/add-access-token.sh"
TA_FULLPATH="$(repack_with_access_token "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
CI_JOB_ID="${CI_JOB_ID:-$(basename $(dirname "$TA_FULLPATH"))}"
TEST_FOLDER="${TEST_FOLDER:-$BUILD_DIR/$CI_JOB_ID}"
mkdir -p "$TEST_FOLDER"
rm -rf "$ADDON_DIR/$REPACKED_TA_NAME"

# Set passthrough env vars config & repackage TA
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/passthrough_env_vars.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'gomemlimit=512MiB' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_debug_config_server=test_notused' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_hec_url=test_notused' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_gateway_url=test_notused' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
cp "$SOURCE_DIR/packaging-scripts/cicd-tests/stdin/passthrough_env_vars.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
tar -C "$ADDON_DIR" -hcz --file "$TA_FULLPATH" "Splunk_TA_otel"


echo "Testing with hot TA $TA_FULLPATH ($ADDON_DIR and $REPACKED_TA_NAME)"

DOCKER_COMPOSE_CONFIG="$SOURCE_DIR/packaging-scripts/cicd-tests/stdin/docker-compose.yml"
REPACKED_TA_NAME=$REPACKED_TA_NAME ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" up --build --force-recreate --wait --detach --timestamps

docker exec -u root stdin-ta-test-stdin-1 /opt/splunk/bin/splunk btool check --debug | grep -qi "Invalid key in stanza" && exit 1

MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if docker exec -u root stdin-ta-test-stdin-1 grep -qi "Everything is ready. Begin running and processing data." /opt/splunk/var/log/splunk/otel.log; then
        break
    else
        if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
            echo "Failed to see success message in otel.log after $MAX_ATTEMPTS attempts."
            docker exec -u root stdin-ta-test-stdin-1 cat /opt/splunk/var/log/splunk/otel.log
            exit 1
        fi
        echo "success message not found in otel.log Retrying in $DELAY seconds"
        ATTEMPT=$((ATTEMPT + 1))
        sleep $DELAY
    fi
done

sample_test="$TEST_FOLDER/splunk-ta-otel-input.xml"
docker exec -u root stdin-ta-test-stdin-1 /opt/splunk/bin/splunk cmd splunkd print-modinput-config Splunk_TA_otel Splunk_TA_otel://Splunk_TA_otel > "$sample_test"
echo "XML which would be provided to our addon can be found at $sample_test"

REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" down
