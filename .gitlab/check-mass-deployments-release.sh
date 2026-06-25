#!/usr/bin/env bash
set -euo pipefail

# Checks whether each mass deployment (Ansible, Chef, Puppet) has unreleased
# changes in its CHANGELOG.md and, for those that do, bumps the minor version,
# stamps the changelog, updates the version file, and opens a single draft PR.
#
# Intended to run on a weekly GitLab CI schedule (Monday early morning).
# GitLab CI variables required: GITHUB_TOKEN, GITHUB_BOT_GPG_KEY, GITHUB_BOT_GPG_KEY_ID
#
# Optional variables:
#   CREATE_PR  Set to "false" to skip PR creation (useful for testing). Default: "true".

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

CREATE_PR="${CREATE_PR:-true}"
if [[ "$CREATE_PR" == "true" ]]; then
  BRANCH="create-pull-request/mass-deployments-release"
  REPO="signalfx/splunk-otel-collector"
  REPO_URL="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${REPO}.git"
fi

# Returns 0 when the ## Unreleased section contains at least one non-blank line.
has_unreleased_content() {
  local changelog="$1"
  awk '/^## Unreleased/{found=1; next} found && /^## /{exit} found && /[^[:space:]]/{print}' "$changelog" | grep -q .
}

# Extracts the version number (without the "v" prefix) from the first versioned
# entry matching ## <prefix>-vX.Y.Z in the changelog.
get_current_version() {
  local changelog="$1"
  local prefix="$2"
  grep -m1 "^## ${prefix}-v" "$changelog" | sed "s/^## ${prefix}-v//"
}

# Increments the minor component and resets the patch to 0 (X.Y.Z -> X.Y+1.0).
bump_minor_version() {
  local version="$1"
  if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "ERROR: expected X.Y.Z version, got: '${version}'" >&2
    exit 1
  fi
  local major minor
  major=$(echo "$version" | cut -d. -f1)
  minor=$(echo "$version" | cut -d. -f2)
  echo "${major}.$((minor + 1)).0"
}

# Replaces the bare ## Unreleased header with ## <prefix>-v<new_version> and
# re-inserts an empty ## Unreleased section above it so future runs can detect
# new entries without requiring a manual edit.
stamp_changelog() {
  local changelog="$1"
  local prefix="$2"
  local new_version="$3"
  sed -i "s/^## Unreleased$/## Unreleased\n\n## ${prefix}-v${new_version}/" "$changelog"
}

# Sources a deployment descriptor, processes its changelog, and appends to the
# changed_deployments array if there are unreleased changes.
# Globals read:  changed_deployments (array, must be declared by caller)
# Globals set:   none
# Args: $1 = path to deployment descriptor (.sh file)
process_deployment() {
  local descriptor="$1"

  # Each descriptor defines: DEPLOYMENT_NAME, DEPLOYMENT_CHANGELOG,
  # DEPLOYMENT_PREFIX, and update_version_file().
  # shellcheck source=/dev/null
  source "$descriptor"

  if has_unreleased_content "$DEPLOYMENT_CHANGELOG"; then
    local current new
    current=$(get_current_version "$DEPLOYMENT_CHANGELOG" "$DEPLOYMENT_PREFIX")
    new=$(bump_minor_version "$current")
    echo ">>> ${DEPLOYMENT_NAME}: ${current} -> ${new}"
    stamp_changelog "$DEPLOYMENT_CHANGELOG" "$DEPLOYMENT_PREFIX" "$new"
    update_version_file "$new"
    changed_deployments+=("${DEPLOYMENT_NAME} ${DEPLOYMENT_PREFIX}-v${new}")
  else
    echo ">>> ${DEPLOYMENT_NAME}: no unreleased changes, skipping"
  fi
}

create_mass_deployments_pr() {
  if [[ "$CREATE_PR" == "true" ]]; then
    echo ">>> Cloning the $REPO repository ..."
    git clone "$REPO_URL" collector-mirror
    cd collector-mirror

    setup_branch "$BRANCH" "$REPO_URL"
  fi

  local changed_deployments=()

  for descriptor in ".gitlab/mass-deployments/"*.sh; do
    process_deployment "$descriptor"
  done

  if [[ ${#changed_deployments[@]} -eq 0 ]]; then
    echo ">>> No mass deployments have unreleased changes. Nothing to do."
    exit 0
  fi

  local names_csv
  names_csv=$(printf '%s, ' "${changed_deployments[@]}")
  names_csv="${names_csv%, }"
  local message="Prepare mass deployment releases: ${names_csv}"

  local body_list
  body_list=$(printf '%s\n' "${changed_deployments[@]}" | sed 's/^/- /')

  echo ">>> Changes detected: ${names_csv}"

  if ! git diff --exit-code >/dev/null 2>&1; then
    if [[ "$CREATE_PR" != "true" ]]; then
      echo ">>> CREATE_PR is not 'true'; skipping PR creation."
      echo ">>> Commit message would be: $message"
      echo ">>> Commit body would be:"
      echo "$body_list"
      return
    fi
    git commit -S -am "$message"
    git push -f "$REPO_URL" "$BRANCH"
    local pr_count
    pr_count="$( get_pr_count "$BRANCH" "$REPO_URL" )"
    if [ "$pr_count" = "0" ]; then
      echo ">>> Creating the PR ..."
      gh pr create \
        --draft \
        --repo "$REPO" \
        --title "$message" \
        --body "$(printf 'Prepare releases for the following mass deployments:\n\n%s' "$body_list")" \
        --label "Skip Changelog" \
        --base main \
        --head "$BRANCH"
    else
      echo ">>> Open PR already exists for $BRANCH, skipping creation."
    fi
  else
    echo ">>> No file changes despite non-empty unreleased sections; nothing to commit."
  fi
}

if [[ "$CREATE_PR" = "true" ]]; then
  setup_gpg
  import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
  setup_git
fi

create_mass_deployments_pr
