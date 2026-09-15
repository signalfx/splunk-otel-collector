#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <release_version> [versions_file]" >&2
  exit 1
fi

RELEASE_VERSION="$1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSIONS_FILE="${2:-${SCRIPT_DIR}/../versions.yaml}"

SET_VERSION=$(awk '/^  splunk-otel-collector:/{found=1} found&&/^    version:/{print $2; exit}' "$VERSIONS_FILE")
if [[ ! "$SET_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-(beta|rc)\.[0-9]+)?$ ]]; then
  echo "Could not read a valid splunk-otel-collector module set version from $VERSIONS_FILE." >&2
  exit 1
fi

if [[ "$SET_VERSION" == "$RELEASE_VERSION" ]]; then
  echo "all"
  exit 0
fi

# A patch release can leave the module set at its current version when only the
# root collector module is being released. Other releases must update and tag
# the complete module set.
if [[ "$RELEASE_VERSION" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  RELEASE_MAJOR="${BASH_REMATCH[1]}"
  RELEASE_MINOR="${BASH_REMATCH[2]}"
  RELEASE_PATCH="${BASH_REMATCH[3]}"

  if [[ "$SET_VERSION" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    SET_MAJOR="${BASH_REMATCH[1]}"
    SET_MINOR="${BASH_REMATCH[2]}"
    SET_PATCH="${BASH_REMATCH[3]}"

    if (( 10#$RELEASE_MAJOR == 10#$SET_MAJOR &&
          10#$RELEASE_MINOR == 10#$SET_MINOR &&
          10#$RELEASE_PATCH > 10#$SET_PATCH )); then
      echo "root"
      exit 0
    fi
  fi
fi

echo "versions.yaml has the splunk-otel-collector module set at $SET_VERSION but release is $RELEASE_VERSION." >&2
echo "Bump the version value in versions.yaml to $RELEASE_VERSION unless this is a newer patch release for the root collector only." >&2
exit 1
