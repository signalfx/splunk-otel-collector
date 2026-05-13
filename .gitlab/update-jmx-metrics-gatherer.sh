#!/usr/bin/env bash
set -euo pipefail

# NOTE: this script is meant to be run on the GitLab CI, it depends on GitLab CI variables
# Based on https://github.com/signalfx/splunk-otel-java/blob/c9134906c84e9a32a974dec4b380453fe1757410/scripts/propagate-version.sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

update_jmx_gatherer_hash() {
  local hash="$1"
  local hash_ldflag_prefix="-X github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver.MetricsGathererHash="

  if ! grep -q '^BUILD_JMX_GATHERER_HASH=' Makefile; then
    echo "ERROR: BUILD_JMX_GATHERER_HASH not found in Makefile" >&2
    exit 1
  fi

  echo ">>> Updating BUILD_JMX_GATHERER_HASH in Makefile ..."
  awk -v hash="${hash}" -v prefix="${hash_ldflag_prefix}" '
    /^BUILD_JMX_GATHERER_HASH=/ {
      print "BUILD_JMX_GATHERER_HASH=" prefix hash
      next
    }
    { print }
  ' Makefile > Makefile.tmp
  mv Makefile.tmp Makefile
}

create_collector_pr() {
  local repo="signalfx/splunk-otel-collector"
  local repo_url="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${repo}.git"
  local branch="create-pull-request/update-jmx-metric-gatherer"
  local message="Update jmx-metric-gatherer to latest"

  echo ">>> Cloning the $repo repository ..."
  git clone "$repo_url" collector-mirror
  cd collector-mirror

  setup_branch "$branch" "$repo_url"

  echo ">>> Getting latest jmx-metric-gatherer release ..."
  release_info="$(gh release view --repo "https://github.com/open-telemetry/opentelemetry-java-contrib" --json tagName,assets --jq '[.tagName, (.assets[] | select(.name == "opentelemetry-jmx-metrics.jar") | .digest | sub("^sha256:"; ""))] | @tsv')"
  tag="$(echo "$release_info" | awk '{print $1}')"
  hash="$(echo "$release_info" | awk '{print $2}')"
  if [[ -n "$tag" && -n "$hash" ]]; then
    echo ">>> Updating jmx-metric-gatherer version to $tag ..."
    echo "$tag" > packaging/jmx-metric-gatherer-release.txt
    update_jmx_gatherer_hash "$hash"
  else
    echo "ERROR: Failed to get latest release tag and/or JMX metrics jar hash from https://github.com/open-telemetry/opentelemetry-java-contrib !" >&2
    exit 1
  fi

  # Only create the PR if there are changes
  if ! git diff --exit-code >/dev/null 2>&1; then
    create_pr_with_changelog "$repo" "$repo_url" "$branch" "$message" \
      "update-jmx-metric-gatherer-${tag}" "packaging" "Update JMX metrics gatherer to ${tag}"
  fi
}

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
create_collector_pr
