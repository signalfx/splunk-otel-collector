#!/bin/bash -eu
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1

source "${SOURCE_DIR}/packaging-scripts/cicd-tests/add-access-token.sh"
TA_FULLPATH="$(repack_with_access_token "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
rm -rf "$ADDON_DIR/$REPACKED_TA_NAME"

# Set passthrough env vars config & repackage TA
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/passthrough_env_vars.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
cp "$SOURCE_DIR/packaging-scripts/technical-addon/envvars/passthrough_env_vars.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
tar -C "$ADDON_DIR" -hcz --file "$TA_FULLPATH" "Splunk_TA_otel"

echo "Testing with hot TA $TA_FULLPATH ($ADDON_DIR and $REPACKED_TA_NAME)"

DOCKER_COMPOSE_CONFIG="$SOURCE_DIR/packaging-scripts/cicd-tests/envvars/docker-compose.yml"
REPACKED_TA_NAME=$REPACKED_TA_NAME ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" up --build --force-recreate --wait --detach --timestamps

docker exec -u root ta-test-envvars-1 /opt/splunk/bin/splunk btool check --debug | grep -qi "Invalid key in stanza" && exit 1

MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if docker exec -u root ta-test-envvars-1 grep -qi "Everything is ready. Begin running and processing data." /opt/splunk/var/log/splunk/otel.log; then
        break
    else
        if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
            echo "Failed to see success message in otel.log after $MAX_ATTEMPTS attempts."
            cat /opt/splunk/var/log/splunk/otel.log
            exit 1
        fi
        echo "sucess message not found in otel.log Retrying in $DELAY seconds"
        ATTEMPT=$((ATTEMPT + 1))
        sleep $DELAY
    fi
done

docker exec -u root ta-test-envvars-1 grep -qi "9092" /opt/splunk/var/log/splunk/otel.log
docker exec -u root ta-test-envvars-1 grep -qi "kafkametrics receiver is working" /opt/splunk/var/log/splunk/otel.log

# Should trap this
REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" down
