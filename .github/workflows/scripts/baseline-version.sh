#!/usr/bin/env bash
# Shared helpers for the baseline module version guard and auto-tagger.
#
# The baseline module is consumed in-repo via a `replace` directive, but is also
# published as a standalone Go module (github.com/signalfx/splunk-otel-collector/baseline)
# that flavours resolve by tag. To keep the published tag drift-free, every change to
# baseline's shipping files must bump baseline/VERSION, and each version is tagged on
# merge. This mirrors the Helm chart-releaser pattern (Chart.yaml version + ct
# --check-version-increment + tag-on-merge).

set -euo pipefail

# baseline_shipping_changed <from-ref> [to-ref]
# Returns 0 (true) if any shipping file differs between the two refs. Shipping
# files are what an external `go get` consumer actually receives; doc-only
# changes under baseline/ do not require a version bump.
baseline_shipping_changed() {
  local from="$1"
  local to="${2:-HEAD}"
  ! git diff --quiet "$from" "$to" -- \
    'baseline/**/*.go' 'baseline/*.go' 'baseline/go.mod' 'baseline/go.sum'
}

# latest_baseline_tag
# Prints the highest baseline/vX.Y.Z tag (version-sorted), or empty if none.
latest_baseline_tag() {
  git tag -l 'baseline/v[0-9]*' --sort=-version:refname | head -n1
}

# read_version
# Prints the version declared in baseline/VERSION (trimmed).
read_version() {
  tr -d '[:space:]' < baseline/VERSION
}

# semver_gt <a> <b>
# Returns 0 if a > b per semver precedence. Inputs are vX.Y.Z[-prerelease].
# Handles the rule `sort -V` gets wrong: a release outranks its own prerelease
# (v0.160.0 > v0.160.0-rc.1), which is the rc->stable transition each cycle.
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

  # Same base: a release (no prerelease) is greater than any prerelease.
  [ -z "$a_pre" ] && return 0
  [ -z "$b_pre" ] && return 1
  [ "$a_pre" != "$b_pre" ] && \
    [ "$(printf '%s\n%s\n' "$a_pre" "$b_pre" | sort -V | tail -n1)" = "$a_pre" ]
}

# next_patch_version <vX.Y.Z[-pre]>
# Prints vX.Y.(Z+1), dropping any prerelease suffix. Used for cherry-pick hints
# and to advance the stable patch when vX.Y.0 is already tagged.
next_patch_version() {
  local base="${1%%-*}"
  if [[ "$base" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    echo "v${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.$(( BASH_REMATCH[3] + 1 ))"
  else
    echo "invalid version: $1" >&2
    return 1
  fi
}

# derive_baseline_version <otel_version> <ref_version>
# Computes the next baseline version from the dep-update target and the latest
# published baseline version (ref_version = latest tag, or current VERSION when
# no tag exists yet).
#   Stable  (otel_version is vX.Y.Z[.*]): baseline tracks upstream major+minor at
#     patch 0 (vX.Y.0); if that is already tagged, advance the patch instead so
#     the version stays strictly monotonic.
#   Nightly (otel_version is main / a commit SHA): increment the rc suffix on the
#     current base; from a GA ref, start rc.1 on the next minor.
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

  # Nightly: bump the rc on the current base, or open rc.1 on the next minor.
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
