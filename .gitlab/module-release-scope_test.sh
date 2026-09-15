#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK_SCRIPT="${SCRIPT_DIR}/module-release-scope.sh"
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT
VERSIONS_FILE="${TEST_DIR}/versions.yaml"

write_version() {
  local version="$1"
  printf 'module-sets:\n  splunk-otel-collector:\n    version: %s\n    modules: []\n' "$version" > "$VERSIONS_FILE"
}

assert_scope() {
  local expected="$1"
  local release_version="$2"
  local set_version="$3"
  local actual

  write_version "$set_version"
  actual=$("$CHECK_SCRIPT" "$release_version" "$VERSIONS_FILE")
  if [[ "$actual" != "$expected" ]]; then
    echo "Expected scope $expected for release $release_version with module set $set_version, got $actual." >&2
    exit 1
  fi
}

assert_rejected() {
  local release_version="$1"
  local set_version="$2"

  write_version "$set_version"
  if "$CHECK_SCRIPT" "$release_version" "$VERSIONS_FILE" >/dev/null 2>&1; then
    echo "Expected release $release_version with module set $set_version to be rejected." >&2
    exit 1
  fi
}

assert_scope all v0.160.0 v0.160.0
assert_scope all v0.161.0-rc.0 v0.161.0-rc.0
assert_scope root v0.160.1 v0.160.0
assert_scope root v0.160.12 v0.160.2

assert_rejected v0.161.0 v0.160.0
assert_rejected v1.160.1 v0.160.0
assert_rejected v0.160.1 v0.160.2
assert_rejected v0.160.1-rc.0 v0.160.0

printf 'module-release-scope tests passed\n'
