#!/bin/bash -eux
# See https://docs.github.com/en/rest/releases/releases?apiVersion=2022-11-28#get-the-latest-release
#curl -s https://api.github.com/repos/signalfx/splunk-otel-collector/releases/latest

#  -H "Authorization: Bearer <YOUR-TOKEN>" \
#curl -L \
#  -H "Accept: application/vnd.github+json" \
#  -H "X-GitHub-Api-Version: 2022-11-28" \
#  https://api.github.com/repos/signalfx/splunk-otel-collector/releases/latest
# Needs jq installed

# Jq exists in brew and in all standard splunk ci-container's

#### First, update otel version

# if auto-detection doesn't work or you need a specific version,
# You can pass SPLUNK_OTEL_VERSION as an environment variable, or
# You can uncomment and set the below line to what's desired.
#SPLUNK_OTEL_VERSION="v0.88.0"
SPLUNK_OTEL_VERSION="${SPLUNK_OTEL_VERSION:-}"
if [ -z "$SPLUNK_OTEL_VERSION" ]; then
    SPLUNK_OTEL_VERSION="$(curl "https://api.github.com/repos/signalfx/splunk-otel-collector/tags" | jq -r '.[0].name')"
fi
echo "updating otel to version $SPLUNK_OTEL_VERSION"
sed -i'.old' "s/^OTEL_COLLECTOR_VERSION?=.*$/OTEL_COLLECTOR_VERSION?=${SPLUNK_OTEL_VERSION#v}/g" "$SOURCE_DIR/Makefile" && rm "$SOURCE_DIR/Makefile.old"
