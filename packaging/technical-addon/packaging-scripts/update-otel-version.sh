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
    SPLUNK_OTEL_VERSION="$(curl "https://api.github.com/repos/signalfx/splunk-otel-collector/tags" | jq -r '[.[].name | select(test("v[0-9]+\\.[0-9]+\\.[0-9]+")) | sub("^v"; "")] | sort_by(split(".") | map(tonumber)) | last | "v" + .')"
fi
echo "updating otel to version $SPLUNK_OTEL_VERSION"
sed -i'.old' "s/^OTEL_COLLECTOR_VERSION?=.*$/OTEL_COLLECTOR_VERSION?=${SPLUNK_OTEL_VERSION#v}/g" "$ADDONS_SOURCE_DIR/Makefile" && rm "$ADDONS_SOURCE_DIR/Makefile.old"
sed -i "s/^EXPECTED_ADDON_VERSION?=.*$/EXPECTED_ADDON_VERSION?=${SPLUNK_OTEL_VERSION#v}/g" "$ADDONS_SOURCE_DIR/cicd-tests/happypath-test.sh"
