#!/bin/bash -eu
set -o pipefail
# Get versions
APP_CONFIG="$ADDONS_SOURCE_DIR/Splunk_TA_otel/default/app.conf"
TA_VERSION="$(sed -n 's/^version = \(.*$\)/\1/p' "$APP_CONFIG" | head -n 1)"
# create new git branch and tag
BRANCH="release/technical-addon/v$TA_VERSION"
REMOTE="origin"
TAG="Splunk_TA_otel/v$TA_VERSION"

git checkout "$BRANCH"
git pull --rebase # in case of any changes made in the PR process
echo "Tagging Collector TA for v$TA_VERSION..."
git tag --annotate --sign --message "Release of Splunk_TA_otel version v$TA_VERSION" "$TAG"


PROMPT="Do you want to push the tag to remote? [y/N] "
USER_INPUT=""
# Check if `read -i` is supported for GNU Bash, else use read command for MacOS Bash
if (help read 2>&1 | grep -q -- "-i"); then
    read -e -i "y" -p "$PROMPT" -r -n 1 USER_INPUT
else
    echo -n "$PROMPT"
    read -r -n 1 USER_INPUT
    echo ""
fi

# Convert input to lowercase and check
if [[ "$USER_INPUT" == "y" ]]; then
    git push "$REMOTE" "$TAG"
    exit "$?"
fi

echo "Tag not pushed. To push this tag to remote, run "
echo "git push $REMOTE '$TAG'"

echo "Check gitlab for status of release pipeline"
