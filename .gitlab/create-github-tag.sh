#!/usr/bin/env bash
set -euo pipefail

# Tags a commit on GitHub and pushes the tags for the whole splunk-otel-collector
# Go module set (root, baseline, and the pkg/* modules), then pushes them.
#
# The module set and its version are declared in versions.yaml and kept in sync
# by `multimod prerelease` as part of the changelog PR, so by release time
# versions.yaml already carries the release version. This script asserts that
# version matches the requested tag, then uses `multimod tag` to create the
# per-module tags (vX.Y.Z, baseline/vX.Y.Z, pkg/*/vX.Y.Z) and pushes them.
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

# Install multimod from the tools module (pinned in internal/tools/go.mod).
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

# Guard: versions.yaml must already carry the release version (synced in the
# changelog PR). This catches an un-synced or mismatched release.
set_version="$(awk '/^  '"$MODULE_SET"':/{f=1} f&&/^    version:/{print $2; exit}' versions.yaml)"
if [[ "$set_version" != "$VERSION_TAG" ]]; then
  echo ">> versions.yaml has module set '$MODULE_SET' at '$set_version' but release is '$VERSION_TAG'." >&2
  echo ">> Run 'multimod prerelease' to sync versions.yaml before releasing." >&2
  exit 1
fi

echo ">>> Creating signed tags for module set $MODULE_SET at $COMMIT_SHA ..."
"$MULTIMOD" tag --module-set-name "$MODULE_SET" --commit-hash "$COMMIT_SHA" --print-tags \
  | grep -E "(^|/)${VERSION_TAG}$" \
  | while IFS= read -r tag; do
      echo ">>> Pushing tag $tag to GitHub ..."
      git push origin "$tag"
    done
