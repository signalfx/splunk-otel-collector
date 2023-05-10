#!/usr/bin/env bash
set -euo pipefail

# NOTE: this script is meant to be run on the GitLab CI, it depends on GitLab CI variables
# Based on https://github.com/signalfx/splunk-otel-java/blob/c9134906c84e9a32a974dec4b380453fe1757410/scripts/propagate-version.sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

create_collector_pr() {
  local repo="signalfx/splunk-otel-collector"
  local repo_url="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${repo}.git"
  local update_deps_branch="create-pull-request/update-deps"
  local message="Update OpenTelemetry Dependencies to latest"

  echo ">>> Cloning the $repo repository ..."
  git clone "$repo_url" collector-mirror
  cd collector-mirror

  # check if the branch exists
  if git ls-remote --exit-code --heads origin "$update_deps_branch"; then
    # get number of open PRs for the branch
    nprs="$( gh pr list --repo "$repo_url" --head "$update_deps_branch" --state open --json id --jq length )"
    if [[ ! "$nprs" =~ ^[0-9]+$ ]]; then
      echo "ERROR: Failed to get PRs for the $update_deps_branch branch!" >&2
      echo "$nprs" >&2
      exit 1
    fi
    if [[ "$nprs" != "0" ]]; then
      echo ">>> The $update_deps_branch branch exists and has $nprs open PR(s)."
      echo ">>> Nothing to do."
      exit 0
    fi
    echo ">>> Resetting the $update_deps_branch branch to main ..."
    git checkout "$update_deps_branch"
    git reset --hard origin/main
  else
    git checkout -b "$update_deps_branch"
  fi

  echo ">>> Updating otel deps to latest ..."
  OTEL_VERSION="latest" ./internal/buildscripts/update-deps
  make for-all CMD='make tidy'

  # Only create the PR if there are changes
  if ! git diff --exit-code >/dev/null 2>&1; then
    git commit -S -am "$message"
    git push -f "$repo_url" "$update_deps_branch"
    echo ">>> Creating the PR ..."
    gh pr create \
      --draft \
      --repo "$repo" \
      --title "$message" \
      --body "$message" \
      --base main \
      --head "$update_deps_branch"
  fi
}

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
create_collector_pr
