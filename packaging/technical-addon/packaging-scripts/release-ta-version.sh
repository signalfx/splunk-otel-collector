#!/bin/bash -eux
# Get versions
APP_CONFIG="$SOURCE_DIR/Splunk_TA_otel/default/app.conf"
TA_VERSION="$(sed -n 's/^version = \(.*$\)/\1/p' "$APP_CONFIG" | head -n 1)"
SPLUNK_OTEL_VERSION="$(sed -n 's/^OTEL_COLLECTOR_VERSION?=\(.*$\)/\1/p' "Makefile" | head -n 1)"

echo "Will create a new release for Splunk_TA_otel @ version $TA_VERSION using splunk otel collector $SPLUNK_OTEL_VERSION"

# create new git branch and tag
BRANCH="release/technical-addon/v$TA_VERSION"
REMOTE="origin"
TAG="Splunk_TA_otel/v$TA_VERSION"

git checkout -B "$BRANCH"
git add "$SOURCE_DIR/Makefile" "$APP_CONFIG"
git commit -m "Updates TA to splunk-otel-collector v$SPLUNK_OTEL_VERSION and marks TA as v$TA_VERSION" || echo "version changes already comitted"


echo "Tagging..."
git tag --annotate --sign --message "Release of Splunk_TA_otel version v$TA_VERSION" "$TAG"
git push --set-upstream "$REMOTE" "$BRANCH"
read -i "Do you want to push the tag to remote too? [y/N]" -r -n 1 USER_INPUT
if [ "$USER_INPUT" == "y" ]; then
    git push "$REMOTE" "$TAG"
    exit "$?"
fi
echo "To push this tag to remote, run "
echo "git push origin '$TAG'"
