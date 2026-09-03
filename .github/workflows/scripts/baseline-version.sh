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
