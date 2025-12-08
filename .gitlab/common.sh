#!/usr/bin/env bash

# Common functions used during the release process on GitLab
# Copied from https://github.com/signalfx/splunk-otel-java/blob/ad61ad1b5fd14249f9e7153607935987b5d7835e/scripts/common.sh

set -e

setup_gpg() {
  echo ">>> Setting GnuPG configuration ..."
  mkdir -p ~/.gnupg
  chmod 700 ~/.gnupg
  cat > ~/.gnupg/gpg.conf <<EOF
no-tty
pinentry-mode loopback
EOF
}

import_gpg_secret_key() {
  local secret_key_contents="$1"

  echo ">>> Importing secret key ..."
  echo "$secret_key_contents" > seckey.gpg
  if (gpg --batch --allow-secret-key-import --import seckey.gpg)
  then
    rm seckey.gpg
  else
    rm seckey.gpg
    exit 1
  fi
}

sign_file() {
  local file="$1"
  echo "$GPG_PASSWORD" | \
    gpg --batch --passphrase-fd 0 --armor --local-user="$GPG_KEY_ID" --detach-sign "$file"
}

setup_git() {
  git config --global user.name release-bot
  git config --global user.email ssg-srv-gh-o11y-gdi@splunk.com
  git config --global gpg.program gpg
  git config --global user.signingKey "$GITHUB_BOT_GPG_KEY_ID"
}

# without the starting 'v'
get_release_version() {
  local release_tag="$1"
  echo "$release_tag" | cut -c2-
}

# 1 from v1.2.3
get_major_version() {
  local release_tag="$1"
  get_release_version "$release_tag" | awk -F'.' '{print $1}'
}
get_minor_version() {
  local release_tag="$1"
  get_release_version "$release_tag" | awk -F'.' '{print $2}'
}
get_patch_version() {
  local release_tag="$1"
  get_release_version "$release_tag" | awk -F'.' '{print $3}'
}

validate_version() {
  local version="$1"
  if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]
  then
    echo "Invalid release version: $version"
    echo "Release version must follow the pattern major.minor.patch, e.g. 1.2.3"
    exit 1
  fi
}

get_pr_count() {
  local branch="$1"
  local repo_url="$2"

  pr_count="$( gh pr list --repo "$repo_url" --head "$branch" --state open --json id --jq length )"
  if [[ ! "$pr_count" =~ ^[0-9]+$ ]]; then
    echo "ERROR: Failed to get PRs for the $branch branch!" >&2
    echo "$pr_count" >&2
    exit 1
  fi

  echo "$pr_count"
}

setup_branch() {
  local branch="$1"
  local repo_url="$2"
  local check_for_pr="${3:-true}"

  # check if the branch exists
  if git ls-remote --exit-code --heads origin "$branch"; then
    if [ "$check_for_pr" = "true" ]; then
      # get number of open PRs for the branch
      pr_count="$( get_pr_count "$branch" "$repo_url" )"
      if [[ "$pr_count" != "0" ]]; then
        echo ">>> The $branch branch exists and has $pr_count open PR(s)."
        echo ">>> Nothing to do."
        exit 0
      fi
    fi
    echo ">>> Resetting the $branch branch to main ..."
    git checkout "$branch"
    git reset --hard origin/main
  else
    git checkout -b "$branch"
  fi
}


# Updates an existing changelog entry with PR number
# Usage: update_changelog_pr_number <changelog_filename> <pr_number>
update_changelog_pr_number() {
  local changelog_file="$1"
  local pr_number="$2"

  if [[ ! -f "$changelog_file" ]]; then
    echo "Warning: Changelog file $changelog_file not found, skipping PR number update." >&2
    return 0
  fi

  if ! command -v yq &> /dev/null; then
    echo "Warning: yq not found, cannot update changelog with PR number" >&2
    return 0
  fi

  echo ">>> Updating changelog with PR #${pr_number} ..."
  
  yq eval -i ".issues += [${pr_number}] | .issues style=\"flow\"" "$changelog_file"
}

# Creates a PR with changelog and automatically updates the changelog with the PR number
# Usage: create_pr_with_changelog <repo> <repo_url> <branch> <message> <changelog_filename> <component> <note> [change_type]
# Example: create_pr_with_changelog "signalfx/splunk-otel-collector" "$repo_url" "$branch" "Update Java agent" "update-javaagent-v1.2.3" "packaging" "Update Java agent to v1.2.3"
create_pr_with_changelog() {
  local repo="$1"
  local repo_url="$2"
  local branch="$3"
  local message="$4"
  local changelog_filename="$5"
  local component="$6"
  local note="$7"
  local change_type="${8:-enhancement}"

  FILENAME="$changelog_filename" \
  COMPONENT="$component" \
  NOTE="$note" \
  CHANGE_TYPE="$change_type" \
  bash "$(dirname "${BASH_SOURCE[0]}")/create-changelog-entry.sh"
  
  local sanitized_filename="${changelog_filename//\//-}"
  sanitized_filename="${sanitized_filename//[^a-zA-Z0-9_-]/-}"
  git add ".chloggen/${sanitized_filename}.yaml"
  
  git commit -S -am "$message"
  git push -f "$repo_url" "$branch"
  
  echo ">>> Creating the PR ..."
  local pr_stdout pr_stderr pr_exit_code
  pr_stderr=$(mktemp)
  
  pr_url=$(gh pr create \
    --draft \
    --repo "$repo" \
    --title "$message" \
    --body "$message" \
    --base main \
    --head "$branch" 2>"$pr_stderr")
  pr_exit_code=$?
  
  if [[ $pr_exit_code -eq 0 ]]; then
    if [[ -n "$pr_url" ]]; then
      pr_number=$(echo "$pr_url" | grep -oE '/pull/[0-9]+' | grep -oE '[0-9]+')
      if [[ -n "$pr_number" ]]; then
        echo ">>> PR #${pr_number} created successfully: $pr_url"
        
        local sanitized_filename="${changelog_filename//\//-}"
        sanitized_filename="${sanitized_filename//[^a-zA-Z0-9_-]/-}"
        update_changelog_pr_number ".chloggen/${sanitized_filename}.yaml" "$pr_number"
        
        git commit -S --amend --no-edit
        git push -f "$repo_url" "$branch"
        echo ">>> Updated PR #${pr_number} with changelog reference"
      else
        echo "Warning: Could not extract PR number from URL: $pr_url" >&2
      fi
    else
      echo "Warning: PR creation succeeded but did not return a URL" >&2
    fi
  else
    echo "ERROR: Failed to create PR (exit code: $pr_exit_code)" >&2
    if [[ -s "$pr_stderr" ]]; then
      echo "Error output from gh pr create:" >&2
      cat "$pr_stderr" >&2
    fi
    rm -f "$pr_stderr"
    return $pr_exit_code
  fi
  
  # Clean up temp file
  rm -f "$pr_stderr"
}
