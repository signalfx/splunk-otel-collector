#!/bin/bash -eu
set -o pipefail

[[ -z "$BUILD_DIR" ]] && echo "BUILD_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1

source "${SOURCE_DIR}/packaging-scripts/cicd-tests/add-access-token.sh"
SPLUNK_APPS_URL="$(repack_with_access_token "foobar" "$BUILD_DIR/out/distribution/Splunk_TA_otel.tgz" | tail -n 1)"
REPACKED_TA_NAME="$(basename "$SPLUNK_APPS_URL")"
ADDON_DIR="$(realpath "$(dirname "$SPLUNK_APPS_URL")")"
echo "Testing with hot TA $SPLUNK_APPS_URL ($ADDON_DIR and $REPACKED_TA_NAME)"
DOCKER_COMPOSE_CONFIG="$SOURCE_DIR/packaging-scripts/cicd-tests/smoketests/docker-compose.yml"
ADDON_DIR="$ADDON_DIR" REPACKED_TA_NAME="$REPACKED_TA_NAME" docker compose --file "$DOCKER_COMPOSE_CONFIG" up --detach --wait


# Should trap this
ADDON_DIR="$ADDON_DIR" REPACKED_TA_NAME="$REPACKED_TA_NAME" docker compose --file "$DOCKER_COMPOSE_CONFIG" down
