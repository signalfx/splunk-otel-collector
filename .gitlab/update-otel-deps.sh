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

  echo ">>> Cloning the $repo repository ..."
  git clone "$repo_url" collector-mirror
  cd collector-mirror

  setup_branch "$BRANCH" "$repo_url"

  echo ">>> Updating otel deps to $OTEL_VERSION ..."
  if [[ "$OTEL_VERSION" == "main" ]]; then
    CORE_VERSION=$(git ls-remote https://github.com/open-telemetry/opentelemetry-collector main | awk '{print $1}')
    CONTRIB_VERSION=$(git ls-remote https://github.com/open-telemetry/opentelemetry-collector-contrib main | awk '{print $1}')
    CORE_VERSION="$CORE_VERSION" CONTRIB_VERSION="$CONTRIB_VERSION" ./internal/buildscripts/update-deps
  else
    OTEL_VERSION="$OTEL_VERSION" ./internal/buildscripts/update-deps
  fi

  # Only create the PR if there are changes
  if ! git diff --exit-code >/dev/null 2>&1; then
    git commit -S -am "$message"
    git push -f "$repo_url" "$BRANCH"
    echo ">>> Creating the PR ..."
    gh pr create \
      --draft \
      --repo "$repo" \
      --title "$message" \
      --body "$message" \
      --base main \
      --head "$BRANCH"
  fi
}

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
create_collector_pr
