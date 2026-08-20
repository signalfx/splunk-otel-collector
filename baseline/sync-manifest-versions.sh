#!/usr/bin/env bash
# Sync builder-config.yaml versions to the resolved versions in go.mod.
#
# The OTel dependency-bump tooling (../update-deps) uses `go get`, which only
# touches go.mod. builder-config.yaml pins the same modules with hardcoded
# versions (otelcol_version + a version on every `- gomod:` line) that would
# otherwise drift. This script rewrites those to match go.mod so the manifest
# stays the source of truth for regeneration.
#
# It changes only version strings; component membership is edited by hand and
# picked up by `go generate`. Run from the baseline module directory.
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${SCRIPT_DIR}"

MANIFEST="builder-config.yaml"
TMP="$( mktemp )"
trap 'rm -f "${TMP}"' EXIT

# otelcol_version is the ocb dist version — the collector release, matching the
# go.opentelemetry.io/collector/otelcol module minus the leading "v".
otelcol_ver="$( go list -m -f '{{.Version}}' go.opentelemetry.io/collector/otelcol )"

while IFS= read -r line; do
  case "${line}" in
    *"otelcol_version:"*)
      # Preserve leading indentation, replace the value.
      printf '%s%s\n' "${line%%otelcol_version:*}" "otelcol_version: ${otelcol_ver#v}"
      ;;
    *"- gomod:"*)
      # Line shape: "  - gomod: <module> <version>". Re-resolve <version> from
      # go.mod for <module>; leave the line untouched if it isn't a dependency.
      mod="$( printf '%s\n' "${line}" | awk '{print $3}' )"
      ver="$( go list -m -f '{{.Version}}' "${mod}" 2>/dev/null || true )"
      if [ -n "${ver}" ]; then
        printf '%s- gomod: %s %s\n' "${line%%- gomod:*}" "${mod}" "${ver}"
      else
        printf '%s\n' "${line}"
      fi
      ;;
    *)
      printf '%s\n' "${line}"
      ;;
  esac
done < "${MANIFEST}" > "${TMP}"

mv "${TMP}" "${MANIFEST}"
trap - EXIT
echo ">>> Synced ${MANIFEST} versions to go.mod (otelcol_version: ${otelcol_ver#v})"
