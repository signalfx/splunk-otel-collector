#!/usr/bin/env bash
set -euo pipefail

# Checks whether each mass deployment (Ansible, Chef, Puppet) has unreleased
# changes in its CHANGELOG.md and, for those that do, bumps the minor version,
# stamps the changelog, updates the version file, and opens a single draft PR.
#
# Intended to run on a weekly GitLab CI schedule (Monday early morning).
# GitLab CI variables required: GITHUB_TOKEN, GITHUB_BOT_GPG_KEY, GITHUB_BOT_GPG_KEY_ID, GPG_KEY_ID, GPG_PASSWORD

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source-path=SCRIPTDIR
source "${SCRIPT_DIR}/common.sh"

ROOT_DIR="${SCRIPT_DIR}/../"
cd "${ROOT_DIR}"

BRANCH="create-pull-request/mass-deployments-release"
REPO="signalfx/splunk-otel-collector"
REPO_URL="https://srv-gh-o11y-gdi:${GITHUB_TOKEN}@github.com/${REPO}.git"

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
  local major minor
  major=$(echo "$version" | cut -d. -f1)
  minor=$(echo "$version" | cut -d. -f2)
  echo "${major}.$((minor + 1)).0"
}

# Replaces the bare ## Unreleased header with ## <prefix>-v<new_version>.
stamp_changelog() {
  local changelog="$1"
  local prefix="$2"
  local new_version="$3"
  sed -i "s/^## Unreleased$/## ${prefix}-v${new_version}/" "$changelog"
}

create_mass_deployments_pr() {
  echo ">>> Cloning the $REPO repository ..."
  git clone "$REPO_URL" collector-mirror
  cd collector-mirror

  setup_branch "$BRANCH" "$REPO_URL"

  local changed_deployments=()

  # --- Ansible ---
  local ansible_changelog="deployments/ansible_collections/signalfx/splunk_otel_collector/CHANGELOG.md"
  if has_unreleased_content "$ansible_changelog"; then
    local ansible_current ansible_new
    ansible_current=$(get_current_version "$ansible_changelog" "ansible")
    ansible_new=$(bump_minor_version "$ansible_current")
    echo ">>> Ansible: $ansible_current -> $ansible_new"
    stamp_changelog "$ansible_changelog" "ansible" "$ansible_new"
    sed -i "s/^version: .*/version: ${ansible_new}/" \
      "deployments/ansible_collections/signalfx/splunk_otel_collector/galaxy.yml"
    changed_deployments+=("Ansible ansible-v${ansible_new}")
  else
    echo ">>> Ansible: no unreleased changes, skipping"
  fi

  # --- Chef ---
  local chef_changelog="deployments/chef/CHANGELOG.md"
  if has_unreleased_content "$chef_changelog"; then
    local chef_current chef_new
    chef_current=$(get_current_version "$chef_changelog" "chef")
    chef_new=$(bump_minor_version "$chef_current")
    echo ">>> Chef: $chef_current -> $chef_new"
    stamp_changelog "$chef_changelog" "chef" "$chef_new"
    sed -i "s/^version '.*'/version '${chef_new}'/" "deployments/chef/metadata.rb"
    changed_deployments+=("Chef chef-v${chef_new}")
  else
    echo ">>> Chef: no unreleased changes, skipping"
  fi

  # --- Puppet ---
  local puppet_changelog="deployments/puppet/CHANGELOG.md"
  if has_unreleased_content "$puppet_changelog"; then
    local puppet_current puppet_new
    puppet_current=$(get_current_version "$puppet_changelog" "puppet")
    puppet_new=$(bump_minor_version "$puppet_current")
    echo ">>> Puppet: $puppet_current -> $puppet_new"
    stamp_changelog "$puppet_changelog" "puppet" "$puppet_new"
    # metadata.json uses standard JSON; target only the top-level "version" field.
    sed -i "s/\"version\": \"[^\"]*\"/\"version\": \"${puppet_new}\"/" "deployments/puppet/metadata.json"
    changed_deployments+=("Puppet puppet-v${puppet_new}")
  else
    echo ">>> Puppet: no unreleased changes, skipping"
  fi

  if [[ ${#changed_deployments[@]} -eq 0 ]]; then
    echo ">>> No mass deployments have unreleased changes. Nothing to do."
    exit 0
  fi

  local names_csv
  names_csv=$(IFS=', '; echo "${changed_deployments[*]}")
  local message="Prepare mass deployment releases: ${names_csv}"

  local body_list
  body_list=$(printf '%s\n' "${changed_deployments[@]}" | sed 's/^/- /')

  echo ">>> Changes detected: ${names_csv}"

  if ! git diff --exit-code >/dev/null 2>&1; then
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

setup_gpg
import_gpg_secret_key "$GITHUB_BOT_GPG_KEY"
setup_git
create_mass_deployments_pr
