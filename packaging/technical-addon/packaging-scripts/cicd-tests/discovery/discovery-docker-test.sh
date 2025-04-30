#!/bin/bash -eu
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$ADDONS_SOURCE_DIR" ]] && echo "ADDONS_SOURCE_DIR not set" && exit 1

source "${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/test-utils.sh"
TA_FULLPATH="$(repack_with_access_token "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$TA_FULLPATH")"
ADDON_DIR="$(realpath "$(dirname "$TA_FULLPATH")")"
rm -rf "$ADDON_DIR/$REPACKED_TA_NAME"

# Set discovery specific config & repackage TA
echo 'discovery=true' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'discovery_properties=$SPLUNK_OTEL_TA_HOME/configs/kafkametrics.discovery.properties.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/docker_observer_without_ssl_kafkametrics_config.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
DISCOVERY_ADDONS_SOURCE_DIR="${ADDONS_SOURCE_DIR}/packaging-scripts/cicd-tests/discovery"
cp "$DISCOVERY_ADDONS_SOURCE_DIR/docker_observer_without_ssl_kafkametrics_config.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
cp "$DISCOVERY_ADDONS_SOURCE_DIR/kafkametrics.discovery.properties.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
tar -C "$ADDON_DIR" -hcz --file "$TA_FULLPATH" "Splunk_TA_otel"

echo "Testing with hot TA $TA_FULLPATH ($ADDON_DIR and $REPACKED_TA_NAME)"

DOCKER_COMPOSE_CONFIG="$ADDONS_SOURCE_DIR/packaging-scripts/cicd-tests/discovery/docker-compose.yml"
REPACKED_TA_NAME=$REPACKED_TA_NAME ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" up --build --force-recreate --wait --detach --timestamps

# If there's an error in the app, you can try manually installing it or modifying files
# Lines are for debugging only, until we get better testing documentation
#docker exec -u splunk discovery-ta-test-discovery-1 cp -r "/tmp/local-tas/Splunk_TA_otel" "/opt/splunk/etc/apps"
#docker exec -u splunk discovery-ta-test-discovery-1 /opt/splunk/bin/splunk restart
#sleep 1m # If restarting splunk for debugging, wait a bit
#docker exec -u root -it discovery-ta-test-discovery-1 bash

docker exec -u root discovery-ta-test-discovery-1 /opt/splunk/bin/splunk btool check --debug | grep -qi "Invalid key in stanza" && exit 1

MAX_ATTEMPTS=6
DELAY=10
ATTEMPT=1
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    if docker exec -u root discovery-ta-test-discovery-1 grep -qi "Everything is ready. Begin running and processing data." /opt/splunk/var/log/splunk/otel.log; then
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

# Ensure command line flags were passed properly
docker exec -u root discovery-ta-test-discovery-1 pgrep -f 'discovery'
docker exec -u root discovery-ta-test-discovery-1 pgrep -f 'discovery-properties'
docker exec -u root discovery-ta-test-discovery-1 pgrep -f 'kafkametrics.discovery.properties.yaml'
docker exec -u root discovery-ta-test-discovery-1 test -d /opt/splunk/etc/apps/Splunk_TA_otel/configs/discovery/config.d.linux

sleep 15s # Give discovery receiver some time to discover things after the collector is up
docker exec -u root discovery-ta-test-discovery-1 grep -qi "9092" /opt/splunk/var/log/splunk/otel.log
docker exec -u root discovery-ta-test-discovery-1 grep -qi "kafkametrics receiver is working" /opt/splunk/var/log/splunk/otel.log

# Should trap this
REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" down
