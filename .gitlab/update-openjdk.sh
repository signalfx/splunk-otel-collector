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
  local branch="create-pull-request/update-openjdk"
  local message="Update Bundled OpenJDK to latest"

  echo ">>> Cloning the $repo repository ..."
  git clone "$repo_url" collector-mirror
  cd collector-mirror

  setup_branch "$branch" "$repo_url"

  echo ">>> Getting latest openjdk release ..."
  tag="$( gh release view --repo "https://github.com/adoptium/temurin11-binaries" --json tagName --jq 'select(.isDraft|not and .isPrelease|not) | .tagName' )"
  if [[ -n "$tag" ]]; then
    version=$( echo "$tag" | sed 's|^jdk-\(.*\)|\1|' | tr '+' '_' )
    if [[ -n "$version" ]]; then
      echo ">>> Updating openjdk version to $version ..."
      sed -i "s|^ARG JDK_VERSION=.*|ARG JDK_VERSION=${version}|" internal/signalfx-agent/bundle/Dockerfile
    else
      echo "ERROR: Failed to get version from tag name '${tag}'!" >&2
      exit 1
    fi
  fi

  # Only create the PR if there are changes
  if ! git diff --exit-code >/dev/null 2>&1; then
    git commit -S -am "$message"
    git push -f "$repo_url" "$branch"
    echo ">>> Creating the PR ..."
    gh pr create \
      --draft \
      --repo "$repo" \
      --title "$message" \
      --body "$message" \
      --base main \
      --head "$branch"
  fi
}

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
create_collector_pr
