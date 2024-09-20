#!/bin/bash -eux
# Get versions
APP_CONFIG="$SOURCE_DIR/Splunk_TA_otel/default/app.conf"
TA_VERSION="$(sed -n 's/^version = \(.*$\)/\1/p' "$APP_CONFIG" | head -n 1)"
SPLUNK_OTEL_VERSION="$(sed -n 's/^OTEL_COLLECTOR_VERSION?=\(.*$\)/\1/p' "Makefile" | head -n 1)"

echo "Will create a new release for $TA_VERSION using splunk otel collector $SPLUNK_OTEL_VERSION"

# create new git branch and tag
git checkout -B "release/v$TA_VERSION"
git add Makefile "$APP_CONFIG"
git commit -m "Updates TA to splunk-otel-collector v$SPLUNK_OTEL_VERSION and marks TA as v$TA_VERSION" || echo "version changes already comitted"


echo "Pushing to remote..."
git tag --annotate --sign --message "Version v$TA_VERSION" "v$TA_VERSION"
git push --set-upstream origin "release/v$TA_VERSION"
git push origin "v$TA_VERSION"
