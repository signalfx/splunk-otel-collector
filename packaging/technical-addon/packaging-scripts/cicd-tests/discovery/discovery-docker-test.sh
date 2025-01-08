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
echo 'discovery_properties=$SPLUNK_OTEL_TA_HOME/configs/kafkametrics.discovery.properties.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'configd=true' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'configd_dir=$SPLUNK_OTEL_TA_HOME/configs/discovery/config.d.linux' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
echo 'splunk_config=$SPLUNK_OTEL_TA_HOME/configs/docker_observer_without_ssl_kafkametrics_config.yaml' >> "$ADDON_DIR/Splunk_TA_otel/local/inputs.conf"
cp "$DISCOVERY_SOURCE_DIR/docker_observer_without_ssl_kafkametrics_config.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
cp "$DISCOVERY_SOURCE_DIR/kafkametrics.discovery.properties.yaml" "$ADDON_DIR/Splunk_TA_otel/configs/"
#tar -czvf "$ADDON_DIR/Splunk_TA_otel_discovery.tgz" "$ADDON_DIR/Splunk_TA_otel"
tar -czvf "$ADDON_DIR/$REPACKED_TA_NAME" "$ADDON_DIR/Splunk_TA_otel"

# Set perms on addon dir
chmod 777 "$ADDON_DIR"
#chmod 777 "$ADDON_DIR/Splunk_TA_otel_discovery.tgz"
chmod 777 "$ADDON_DIR/$REPACKED_TA_NAME"

# Set env vars to be passed into docker compose
DOCKER_COMPOSE_CONFIG="$SOURCE_DIR/packaging-scripts/cicd-tests/discovery/docker-compose.yml"
KAFKA_DOCKER_COMPOSE_PATH="$SOURCE_DIR/../../docker/docker-compose.yml"
export DISCOVERY_LOGS_DIR="$BUILD_DIR/tests/discovery/discovery-logs"
export KAFKA_LOGS_DIR="$BUILD_DIR/tests/discovery/discovery-logs"
export SPLUNK_LOGS_DIR="$BUILD_DIR/tests/discovery/splunklogs"
export SPLUNK_APPS_DIR="$BUILD_DIR/tests/discovery/splunkapps"
mkdir -p --mode 777 "$DISCOVERY_LOGS_DIR"
mkdir -p --mode 777 "$SPLUNK_LOGS_DIR"
mkdir -p --mode 777 "$SPLUNK_APPS_DIR"
mkdir -p --mode 777 "$KAFKA_LOGS_DIR"
echo "Will write discovery test logs to $DISCOVERY_LOGS_DIR"

#REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR DISCOVERY_LOGS_DIR="$DISCOVERY_LOGS_DIR" docker compose  --file "$DOCKER_COMPOSE_CONFIG" config
REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR DISCOVERY_LOGS_DIR="$DISCOVERY_LOGS_DIR" docker compose --file "$DOCKER_COMPOSE_CONFIG" up --build --wait --detach
#REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR DISCOVERY_LOGS_DIR="$DISCOVERY_LOGS_DIR" docker compose  --file "$DOCKER_COMPOSE_CONFIG" up --build
#docker exec -u splunk discovery-ta-test-discovery-1 /opt/splunk/bin/splunk install app "/addon-dir/$REPACKED_TA_NAME" -auth 'admin:Chang3d!'
# I can't seem to find a way to directly install it to the apps folder after a bunch of trying, possibly due to bug in config.  So, copy it and restart.
docker exec -u splunk discovery-ta-test-discovery-1 cp -r "/addon-dir/Splunk_TA_otel" "/opt/splunk/etc/apps"


# Normally, we test discovery mode by running the "raw" binary in our tests
# However, here our binary via the TA is running in a splunk docker instance
# The splunk image does not have docker installed by default
# Thus, either need to install and enable docker, 
# Or need to bridge networks for discovery.
# Given lack of a package manager in splunk docker images, it's likely best to just use host networking.

#docker exec -u root discovery-ta-test-discovery-1 yum install -y docker
#docker exec -u root discovery-ta-test-discovery-1 systemctl restart docker
docker exec -u splunk discovery-ta-test-discovery-1 /opt/splunk/bin/splunk restart
# TODO delete this, used for testing/inspection
#docker exec -u root -it discovery-ta-test-discovery-1 cat /opt/splunk/var/log/splunk/otel.log
docker exec -u root -it discovery-ta-test-discovery-1 bash
# Wait for this to come online, *then* check splunk docker logs
#JVM_OPTS="" docker compose --file "$KAFKA_DOCKER_COMPOSE_PATH" --profile integration-test-ta-discovery up -d --wait --build --quiet-pull

# Check logs on host
#docker exec 

# Should trap these
REPACKED_TA_NAME=$REPACKED_TA_NAME BUILD_DIR=$BUILD_DIR ADDON_DIR=$ADDON_DIR DISCOVERY_LOGS_DIR="$DISCOVERY_LOGS_DIR" docker compose --file "$DOCKER_COMPOSE_CONFIG" down
#JVM_OPTS="" docker compose --file "$KAFKA_DOCKER_COMPOSE_PATH" --profile integration-test-ta-discovery down
