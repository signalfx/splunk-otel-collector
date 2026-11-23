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
# Usage: ./create-github-tag.sh <version_tag> <commit_sha> [release_snapshot]

if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "Usage: $0 <version_tag> <commit_sha> [release_snapshot]" >&2
  exit 1
fi

VERSION_TAG="$1"
COMMIT_SHA="$2"
RELEASE_SNAPSHOT="${3:-}"
MODULE_SET="splunk-otel-collector"
REPO="signalfx/splunk-otel-collector"
REPO_URL="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${REPO}.git"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=.gitlab/common.sh
source "$SCRIPT_DIR/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

echo ">>> Cloning $REPO ..."
git clone --no-checkout "$REPO_URL" repo-tmp
cd repo-tmp
git fetch --tags origin

if git show-ref --verify --quiet "refs/tags/${VERSION_TAG}"; then
  tagged_commit="$(git rev-parse --verify "refs/tags/${VERSION_TAG}^{commit}")"
  if [[ "$tagged_commit" != "$COMMIT_SHA" ]]; then
    echo "$VERSION_TAG already points to $tagged_commit instead of $COMMIT_SHA" >&2
    exit 1
  fi
  if [[ -n "$RELEASE_SNAPSHOT" ]]; then
    echo "$VERSION_TAG appeared during publication. Verify the tag by running recover-stable-release in the original child pipeline." >&2
    exit 1
  fi
  # The module set's tags are pushed atomically, so the original job can resume.
  echo "Module tags for $VERSION_TAG are already published at $COMMIT_SHA"
  exit 0
fi

if ! command -v gpg > /dev/null; then
  apt-get update
  apt-get install -y gnupg
fi
setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
git checkout "$COMMIT_SHA"

echo ">>> Creating signed tags for module set $MODULE_SET ($VERSION_TAG) at $COMMIT_SHA ..."
tags="$( multimod tag --module-set-name "$MODULE_SET" --commit-hash "$COMMIT_SHA" --print-tags | sort -u )"
if [[ -n "$RELEASE_SNAPSHOT" ]]; then
  # Update the local root tag before the atomic push of the module tags.
  if ! grep -Fxq "$VERSION_TAG" <<< "$tags"; then
    echo "Module tags did not include the release tag $VERSION_TAG" >&2
    exit 1
  fi
  tag_subject="$(git for-each-ref --format='%(contents:subject)' "refs/tags/${VERSION_TAG}")"
  git tag -f -s -m "$tag_subject" -m "Collector-Release-Snapshot: $RELEASE_SNAPSHOT" "$VERSION_TAG" "$COMMIT_SHA"
fi

echo ">>> Pushing tags to GitHub: $tags"
# Word-split $tags into args; module tags never contain spaces. --atomic so the
# whole module set is published or none of it is.
# shellcheck disable=SC2086
git push --atomic origin $tags
