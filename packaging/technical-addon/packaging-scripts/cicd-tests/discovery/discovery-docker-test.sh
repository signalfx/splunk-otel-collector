#!/bin/bash -eux
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1

source "${SOURCE_DIR}/packaging-scripts/cicd-tests/add-access-token.sh"
BUILD_DIR="$(realpath "$BUILD_DIR")"
DISCOVERY_SOURCE_DIR="${SOURCE_DIR}/packaging-scripts/cicd-tests/discovery"
SPLUNK_APPS_URL="$(repack_with_access_token "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$SPLUNK_APPS_URL")"
ADDON_DIR="$(realpath "$(dirname "$SPLUNK_APPS_URL")")"
echo "Testing with hot TA $SPLUNK_APPS_URL ($ADDON_DIR and $REPACKED_TA_NAME)"

# Set discovery specific config & repackage TA
echo 'discovery=true' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_discovery_properties=$SPLUNK_OTEL_TA_HOME/configs/kafkametrics.discovery.properties.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'configd=true' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'configd_dir=$SPLUNK_OTEL_TA_HOME/configs/discovery/config.d.linux' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/docker_observer_without_ssl_kafkametrics_config.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
cp "$DISCOVERY_SOURCE_DIR/docker_observer_without_ssl_kafkametrics_config.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
cp "$DISCOVERY_SOURCE_DIR/kafkametrics.discovery.properties.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
tar -czvf "$ADDON_DIR/$REPACKED_TA_NAME" "$ADDON_DIR/Splunk_TA_otel"

# Set perms on addon dir
chmod 777 "$ADDON_DIR"
chmod 777 "$ADDON_DIR/$REPACKED_TA_NAME"

DOCKER_COMPOSE_CONFIG="$SOURCE_DIR/packaging-scripts/cicd-tests/discovery/docker-compose.yml"
REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" up --build --force-recreate --wait --detach

# These seem required for some reason... never seems to work when directly starting it up via command line args or install command
docker exec -u splunk discovery-ta-test-discovery-1 cp -r "/addon-dir/Splunk_TA_otel" "/opt/splunk/etc/apps"
docker exec -u splunk discovery-ta-test-discovery-1 /opt/splunk/bin/splunk restart
# If there's an error in the app, you can try manually installing it or modifying files
# Lines are for debugging only, until we get better testing documentation
#docker exec -u root -it discovery-ta-test-discovery-1 bash

docker exec -u root -it discovery-ta-test-discovery-1 /opt/splunk/bin/splunk btool check --debug | grep -qi "Invalid key in stanza" && exit 1
sleep 1m # need to figure out better way to do wait
docker exec -u root -it discovery-ta-test-discovery-1 grep -qi "9092" /opt/splunk/var/log/splunk/otel.log
docker exec -u root -it discovery-ta-test-discovery-1 pgrep -f 'discovery'
docker exec -u root -it discovery-ta-test-discovery-1 pgrep -f 'discovery-properties'
docker exec -u root -it discovery-ta-test-discovery-1 pgrep -f 'kafkametrics.discovery.properties.yaml'
docker exec -u root -it discovery-ta-test-discovery-1 grep -i "kafkametrics receiver is working" /opt/splunk/var/log/splunk/otel.log

# Should trap this
REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR docker compose --file "$DOCKER_COMPOSE_CONFIG" down
