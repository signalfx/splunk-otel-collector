#!/usr/bin/env bash
set -euo pipefail

# Tags a commit on GitHub for the whole splunk-otel-collector Go module set
# (root, baseline, and the pkg/* modules) and pushes the tags.
#
# The module set and its version live in versions.yaml, already validated
# against the release by the manual_release_trigger job. This uses `multimod
# tag` to create the signed per-module tags (vX.Y.Z, baseline/vX.Y.Z,
# pkg/*/vX.Y.Z) and pushes them.
#
# Usage: ./create-github-tag.sh <version_tag> <commit_sha>

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <version_tag> <commit_sha>" >&2
  exit 1
fi

VERSION_TAG="$1"
COMMIT_SHA="$2"
MODULE_SET="splunk-otel-collector"
REPO="signalfx/splunk-otel-collector"
REPO_URL="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${REPO}.git"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$SCRIPT_DIR/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

echo ">>> Installing multimod ..."
( cd ./internal/tools && go install go.opentelemetry.io/build-tools/multimod )
MULTIMOD="$(go env GOPATH)/bin/multimod"

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git

echo ">>> Cloning $REPO ..."
git clone --no-checkout "$REPO_URL" repo-tmp
cd repo-tmp
git fetch origin
git checkout "$COMMIT_SHA"

echo ">>> Creating signed tags for module set $MODULE_SET ($VERSION_TAG) at $COMMIT_SHA ..."
"$MULTIMOD" tag --module-set-name "$MODULE_SET" --commit-hash "$COMMIT_SHA" --print-tags \
  | sort -u \
  | while IFS= read -r tag; do
      echo ">>> Pushing tag $tag to GitHub ..."
      git push origin "$tag"
    done
