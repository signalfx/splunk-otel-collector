#!/usr/bin/env bash
set -euo pipefail

# Purpose: Automates the creation of changelog entries based on input parameters.
#
# Required Parameters:
#   - FILENAME: Name of the changelog file to create (without .yaml extension)
#   - COMPONENT: Component affected by the change (e.g., 'collector', 'packaging')
#   - NOTE: Brief description of the change
#
# Optional Parameters:
#   - CHANGE_TYPE: Type of change (defaults to 'enhancement')
#                  Options: 'breaking', 'deprecation', 'new_component', 'enhancement', 'bug_fix'
#   - ISSUES: Space-separated list of issue numbers (e.g., "123 456")
#   - SUBTEXT: Additional information for the changelog entry (multiline supported)
#
# Examples:
#   FILENAME="fix-bug" COMPONENT="receiver" NOTE="Fix memory leak" CHANGE_TYPE="bug_fix" ./create-changelog-entry.sh
#   FILENAME="update-deps" COMPONENT="packaging" NOTE="Update to v1.2.3" ./create-changelog-entry.sh

if ! command -v yq &> /dev/null; then
  echo "Error: yq is not installed. Please install yq: https://github.com/mikefarah/yq" >&2
  echo "  On macOS: brew install yq" >&2
  echo "  On Linux: Download from https://github.com/mikefarah/yq/releases" >&2
  exit 1
fi

# ---- Changelog Entry Validation ----
if [[ -z "${FILENAME:-}" ]]; then
  echo "Error: FILENAME is required" >&2
  echo "Usage: FILENAME=<name> COMPONENT=<component> NOTE=<note> $0" >&2
  exit 1
fi

if [[ -z "${COMPONENT:-}" ]]; then
  echo "Error: COMPONENT is required" >&2
  echo "Usage: FILENAME=<name> COMPONENT=<component> NOTE=<note> $0" >&2
  exit 1
fi

if [[ -z "${NOTE:-}" ]]; then
  echo "Error: NOTE is required" >&2
  echo "Usage: FILENAME=<name> COMPONENT=<component> NOTE=<note> $0" >&2
  exit 1
fi

CHANGE_TYPE="${CHANGE_TYPE:-enhancement}"
ISSUES="${ISSUES:-}"
SUBTEXT="${SUBTEXT:-}"

# Sanitize filename: replace slashes and special chars with dashes
FILENAME="${FILENAME//\//-}"
FILENAME="${FILENAME//[^a-zA-Z0-9_-]/-}"
changelog_file=".chloggen/${FILENAME}.yaml"

if [ ! -d ".chloggen" ]; then
  echo "Error: .chloggen/ directory not found. Ensure it exists." >&2
  exit 1
fi

if [ -f "$changelog_file" ]; then
  echo "Warning: Changelog entry ${FILENAME}.yaml already exists. Overwriting."
fi

echo "Creating changelog entry: $changelog_file"

yq eval -n '{
  "change_type": env(CHANGE_TYPE),
  "component": env(COMPONENT),
  "note": env(NOTE),
  "issues": [],
  "subtext": null
}' > "$changelog_file"

# Update issues if provided - convert space-separated string to YAML array
if [[ -n "$ISSUES" ]]; then
  issues_json="[$(echo "$ISSUES" | sed 's/ /, /g')]"
  yq eval -i ".issues = $issues_json" "$changelog_file"
fi

# Update subtext if provided
if [[ -n "$SUBTEXT" ]]; then
  SUBTEXT="$SUBTEXT" yq eval -i '.subtext = strenv(SUBTEXT)' "$changelog_file"
fi

yq eval -i '.issues style="flow"' "$changelog_file"

echo "Changelog entry ${FILENAME}.yaml has been created."
exit 0
