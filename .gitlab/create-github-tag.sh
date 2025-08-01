#!/usr/bin/env bash
set -euo pipefail

# This script handles tagging a commit on GitHub and pushing the tag.
# Usage: ./tag-on-github.sh <version_tag> <commit_sha>

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <version_tag> <commit_sha>" >&2
  exit 1
fi

VERSION_TAG="$1"
COMMIT_SHA="$2"
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

echo ">>> Creating signed tag $VERSION_TAG for $COMMIT_SHA ..."
git tag -s "$VERSION_TAG" "$COMMIT_SHA" -m "Release $VERSION_TAG"

echo ">>> Pushing tag $VERSION_TAG to GitHub ..."
git push origin "$VERSION_TAG"
