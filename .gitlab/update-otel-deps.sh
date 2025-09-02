#!/usr/bin/env bash
set -euo pipefail

# NOTE: this script is meant to be run on the GitLab CI, it depends on GitLab CI variables
# Based on https://github.com/signalfx/splunk-otel-java/blob/c9134906c84e9a32a974dec4b380453fe1757410/scripts/propagate-version.sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
OTEL_VERSION="${1:-latest}"
BRANCH="${2:-create-pull-request/update-deps}"

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

create_collector_pr() {
  local repo="signalfx/splunk-otel-collector"
  local repo_url="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${repo}.git"
  local message="Update OpenTelemetry Dependencies to $OTEL_VERSION"
  local check_for_pr="true"
  if [ "$OTEL_VERSION" = "main" ]; then
    check_for_pr="false"
  fi

  echo ">>> Cloning the $repo repository ..."
  git clone "$repo_url" collector-mirror
  cd collector-mirror

  setup_branch "$BRANCH" "$repo_url" "$check_for_pr"

  echo ">>> Updating otel deps to $OTEL_VERSION ..."
  OTEL_VERSION="$OTEL_VERSION" ./update-deps

  # Only create the PR if there are changes
  if ! git diff --exit-code >/dev/null 2>&1; then
    git commit -S -am "$message"
    git push -f "$repo_url" "$BRANCH"
    pr_count="$( get_pr_count "$BRANCH" "$repo_url" )"
    if [ "$pr_count" = "0" ]; then
      echo ">>> Creating the PR ..."
      gh pr create \
        --draft \
        --repo "$repo" \
        --title "$message" \
        --body "$message" \
        --label "Skip Changelog"
        --base main \
        --head "$BRANCH"
    fi
  fi
}

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
create_collector_pr
