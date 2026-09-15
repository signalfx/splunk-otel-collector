#!/usr/bin/env bash
set -euo pipefail

# Tags a commit on GitHub and pushes the tags. When versions.yaml matches the
# release, this tags the whole splunk-otel-collector Go module set (root,
# baseline, and the pkg/* modules). A newer patch release in the same series
# can leave versions.yaml unchanged and tag only the root collector module.
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

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git

echo ">>> Cloning $REPO ..."
git clone --no-checkout "$REPO_URL" repo-tmp
cd repo-tmp
git fetch origin
git checkout "$COMMIT_SHA"

MODULE_RELEASE_SCOPE=$(./.gitlab/module-release-scope.sh "$VERSION_TAG")
if [[ "$MODULE_RELEASE_SCOPE" == "all" ]]; then
  echo ">>> Creating signed tags for module set $MODULE_SET ($VERSION_TAG) at $COMMIT_SHA ..."
  tags="$( multimod tag --module-set-name "$MODULE_SET" --commit-hash "$COMMIT_SHA" --print-tags | sort -u )"
else
  echo ">>> Creating signed root module tag $VERSION_TAG at $COMMIT_SHA ..."
  git tag -s "$VERSION_TAG" "$COMMIT_SHA" -m "Release $VERSION_TAG"
  tags="$VERSION_TAG"
fi

echo ">>> Pushing tags to GitHub: $tags"
# Word-split $tags into args; module tags never contain spaces. --atomic so the
# whole module set is published or none of it is.
# shellcheck disable=SC2086
git push --atomic origin $tags
