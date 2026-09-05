#!/usr/bin/env bash
# Shared helpers for the baseline version guard and auto-tagger.

set -euo pipefail

# True if any file an external consumer receives differs between the two refs.
baseline_shipping_changed() {
  local from="$1"
  local to="${2:-HEAD}"
  ! git diff --quiet "$from" "$to" -- \
    'baseline/**/*.go' 'baseline/*.go' 'baseline/go.mod' 'baseline/go.sum'
}

latest_baseline_tag() {
  git tag -l 'baseline/v[0-9]*' --sort=-version:refname | head -n1
}

read_version() {
  tr -d '[:space:]' < baseline/VERSION
}

# True if a > b per semver, including release > its own prerelease.
semver_gt() {
  local a="$1" b="$2"
  [ "$a" = "$b" ] && return 1

  local a_base="${a%%-*}" b_base="${b%%-*}"
  local a_pre="" b_pre=""
  [ "$a" != "$a_base" ] && a_pre="${a#*-}"
  [ "$b" != "$b_base" ] && b_pre="${b#*-}"

  if [ "$a_base" != "$b_base" ]; then
    [ "$(printf '%s\n%s\n' "$a_base" "$b_base" | sort -V | tail -n1)" = "$a_base" ]
    return
  fi

  [ -z "$a_pre" ] && return 0
  [ -z "$b_pre" ] && return 1
  [ "$a_pre" != "$b_pre" ] && \
    [ "$(printf '%s\n%s\n' "$a_pre" "$b_pre" | sort -V | tail -n1)" = "$a_pre" ]
}

# vX.Y.(Z+1), dropping any prerelease suffix.
next_patch_version() {
  local base="${1%%-*}"
  if [[ "$base" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    echo "v${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.$(( BASH_REMATCH[3] + 1 ))"
  else
    echo "invalid version: $1" >&2
    return 1
  fi
}

# Next baseline version from the dep-update target and the latest published version.
# Stable (otel vX.Y.Z): vX.Y.0, or next patch if already tagged.
# Nightly (otel main/SHA): increment the rc, or open rc.1 on the next minor.
derive_baseline_version() {
  local otel="$1" ref="$2"
  if [[ "$otel" =~ ^v([0-9]+)\.([0-9]+)\.[0-9]+ ]]; then
    local candidate="v${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.0"
    if [ -z "$ref" ] || semver_gt "$candidate" "$ref"; then
      echo "$candidate"
    else
      next_patch_version "$ref"
    fi
    return
  fi

  local base="${ref%%-*}" pre=""
  [ "$ref" != "$base" ] && pre="${ref#*-}"
  if [[ "$pre" =~ ^rc\.([0-9]+)$ ]]; then
    echo "${base}-rc.$(( BASH_REMATCH[1] + 1 ))"
  elif [[ "$base" =~ ^v([0-9]+)\.([0-9]+)\.[0-9]+$ ]]; then
    echo "v${BASH_REMATCH[1]}.$(( BASH_REMATCH[2] + 1 )).0-rc.1"
  else
    echo "invalid ref version: $ref" >&2
    return 1
  fi
}
