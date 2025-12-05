#!/bin/bash -eu
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$ADDONS_SOURCE_DIR" ]] && echo "ADDONS_SOURCE_DIR not set" && exit 1

source "${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
TA_FULLPATH="$(repack_with_test_config "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
export ADDON_DIR
export REPACKED_TA_NAME
echo "Testing with hot TA $TA_FULLPATH ($ADDON_DIR and $REPACKED_TA_NAME)"

if [ ! -f "$TA_FULLPATH" ]; then
    echo "ERROR: TA archive not found at $TA_FULLPATH"
    exit 1
fi
echo "TA archive verified: $(ls -lh "$TA_FULLPATH")"
DOCKER_COMPOSE_CONFIG="$ADDONS_SOURCE_DIR/packaging-scripts/cicd-tests/smoketests/docker-compose.yml"
if ! docker compose --file "$DOCKER_COMPOSE_CONFIG" up --quiet-pull --detach --wait --build --force-recreate --timestamps; then
    echo "ERROR: Docker compose failed to start. Capturing logs..."
    docker compose --file "$DOCKER_COMPOSE_CONFIG" logs
    docker compose --file "$DOCKER_COMPOSE_CONFIG" down
    exit 1
fi

# If there's an error in the app, you can try manually installing it or modifying files
# Lines are for debugging only, until we get better testing documentation
#docker exec -u splunk smoketests-so1-1 cp -r "/tmp/local-tas/Splunk_TA_otel" "/opt/splunk/etc/apps"
#docker exec -u splunk smoketests-so1-1 /opt/splunk/bin/splunk restart
#sleep 1m # If restarting splunk for debugging, wait a bit 

docker exec -u root smoketests-so1-1 /opt/splunk/bin/splunk btool check --debug | grep -qi "Invalid key in stanza" && exit 1

MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if docker exec -u root smoketests-so1-1 grep "Everything is ready. Begin running and processing data." /opt/splunk/var/log/splunk/otel.log; then
        break
    else
        if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
            echo "Failed to see success message in otel.log after $MAX_ATTEMPTS attempts."
            exit 1
        fi
        echo "sucess message not found in otel.log Retrying in $DELAY seconds" 
        ATTEMPT=$((ATTEMPT + 1))
        sleep $DELAY
    fi
done

# Should trap this
docker compose --file "$DOCKER_COMPOSE_CONFIG" down
