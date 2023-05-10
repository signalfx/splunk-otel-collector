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
