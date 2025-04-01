#!/usr/bin/env bash
set -euo pipefail

# NOTE: Triggers Github workflows to create PRs to update the downstream Helm chart.
# - Repo: https://github.com/signalfx/splunk-otel-collector-chart
# - Opens PRs to update instrumentation library dependencies.
# - Opens a PR to prepare a new chart release after a collector release.
# - This script uses the GitHub bot from this repo to trigger downstream chart PRs, ensuring GitHub
#   checks/tests run automatically with these PRs. Without this, maintainers would have to manually
#   reopen chart PRs to trigger the PR checks/tests. In the future, the chart repo may have its own
#   GitLab mirror and Github bot integration.

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

run_chart_github_workflows() {
  local repo="signalfx/splunk-otel-collector-chart"
  local repo_url="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${repo}.git"
  local branch="main"
  local workflows=(
    "update_chart_dependencies.yaml"
    "update_docker_images.yaml"
    "release_drafter.yaml"
  )

  for workflow in "${workflows[@]}"; do
    echo "Triggering workflow: $workflow"
    if gh workflow run "$workflow" --repo "$repo_url" --ref "$branch"; then
      echo "Triggered: $workflow"
    else
      echo "Failed to trigger: $workflow — continuing..."
    fi
  done
}

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
run_chart_github_workflows
