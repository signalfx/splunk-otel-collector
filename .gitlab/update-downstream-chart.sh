#!/usr/bin/env bash
set -euo pipefail

# NOTE: this script is meant to be run on the GitLab CI, it depends on GitLab CI variables
# This script triggers a predefined set of GitHub Actions workflows in the signalfx/splunk-otel-collector
# repository on the main branch. It uses the GitHub CLI (gh) to run each workflow sequentially and exits
# early if any workflow fails to trigger.

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
OTEL_VERSION="${1:-latest}"
BRANCH="${2:-create-pull-request/update-downstream-chart}"
REPO="signalfx/splunk-otel-collector"
REF="main"
WORKFLOWS=(
  "update_chart_dependencies.yaml"
  "update_docker_images.yaml"
  "release_drafter.yaml"
)

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git

for WORKFLOW in "${WORKFLOWS[@]}"; do
  echo "Triggering workflow: $WORKFLOW"
  if gh workflow run "$WORKFLOW" --repo "$REPO" --ref "$REF"; then
    echo "Triggered: $WORKFLOW"
  else
    echo "Failed to trigger: $WORKFLOW — continuing..."
  fi
done

echo "Workflow triggering complete (some may have failed)."
