#!/usr/bin/env bash
# update-go-version.sh
#
# Updates the Go version across the repository to the specified version.
# Usage: ./update-go-version.sh <new-go-version>
# Example: ./update-go-version.sh 1.25.8
#
# This script updates Go version references in:
#   - .gitlab-ci.yml (GO_VERSION variable)
#   - .github/*.yml and *.yaml (GO_VERSION variable)
#   - All go.mod files (go directive)
#   - Dockerfiles with ARG GO_VERSION=...
#   - Dockerfiles with FROM golang:<version>... (suffixes like -bullseye are preserved)
#   - YAML files with golang:<version>... image references
#
# Note: The script uses GNU sed on Linux and BSD sed on macOS.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage: $0 <new-go-version>"
  echo "  <new-go-version>  The Go minor version to update to (e.g. 1.25.8)"
  echo ""
  echo "Example:"
  echo "  $0 1.25.8"
  exit 1
}

if [[ $# -ne 1 ]]; then
  usage
fi

NEW_VERSION="$1"

# Validate version format (e.g. 1.25.8)
if ! echo "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "ERROR: Invalid Go version format: '$NEW_VERSION'"
  echo "Expected format: X.Y.Z (e.g. 1.25.8)"
  exit 1
fi

# Detect sed in-place flag (macOS vs Linux)
if [[ "$(uname)" == "Darwin" ]]; then
  SED_INPLACE=(sed -i '')
else
  SED_INPLACE=(sed -i)
fi

UPDATED_FILES=()

# --- 1. Update GO_VERSION variable in CI/CD config files ---
# Handles .gitlab-ci.yml and .github/*.yml and *.yaml files
while IFS= read -r -d '' cifile; do
  if grep -qE 'GO_VERSION:[[:space:]]*[0-9]+\.[0-9]+\.[0-9]+' "$cifile"; then
    "${SED_INPLACE[@]}" -E "s/(GO_VERSION:[[:space:]]*)[0-9]+\.[0-9]+\.[0-9]+/\1${NEW_VERSION}/g" "$cifile"
    UPDATED_FILES+=("$cifile")
  fi
done < <(find "$REPO_ROOT" \( -name ".gitlab-ci.yml" -o -path "*/.github/*.yml" -o -path "*/.github/*.yaml" \) -not -path "*/vendor/*" -print0)

# --- 2. Update all go.mod files (go directive) ---
# Excludes vendor directories
while IFS= read -r -d '' gomod; do
  if grep -qE '^go [0-9]+\.[0-9]+\.[0-9]+' "$gomod"; then
    "${SED_INPLACE[@]}" -E "s/^go [0-9]+\.[0-9]+\.[0-9]+/go ${NEW_VERSION}/" "$gomod"
    UPDATED_FILES+=("$gomod")
  fi
done < <(find "$REPO_ROOT" -name "go.mod" -not -path "*/vendor/*" -print0)

# --- 3. Update Dockerfiles with ARG GO_VERSION=<version> ---
while IFS= read -r -d '' dockerfile; do
  if grep -qE 'ARG GO_VERSION=[0-9]+\.[0-9]+\.[0-9]+' "$dockerfile"; then
    "${SED_INPLACE[@]}" -E "s/(ARG GO_VERSION=)[0-9]+\.[0-9]+\.[0-9]+/\1${NEW_VERSION}/" "$dockerfile"
    UPDATED_FILES+=("$dockerfile")
  fi
done < <(find "$REPO_ROOT" -name "Dockerfile*" -not -path "*/vendor/*" -print0)

# --- 4. Update Dockerfiles with FROM golang:<version>... ---
# Matches patterns like:
#   FROM golang:1.25.7-bullseye
#   FROM golang:1.25.7-alpine
#   FROM golang:1.25.7
#   FROM --platform=${IMAGE_PLATFORM} golang:1.25.7 as golang
# Only the X.Y.Z portion is replaced; suffixes like -bullseye are preserved.
# Uses POSIX character classes for macOS BSD sed compatibility.
while IFS= read -r -d '' dockerfile; do
  if grep -qE 'FROM[[:space:]]+(--platform=[^[:space:]]+[[:space:]]+)?golang:[0-9]+\.[0-9]+\.[0-9]+' "$dockerfile"; then
    "${SED_INPLACE[@]}" -E "s/(FROM[[:space:]]+(--platform=[^[:space:]]+[[:space:]]+)?golang:)[0-9]+\.[0-9]+\.[0-9]+/\1${NEW_VERSION}/" "$dockerfile"
    UPDATED_FILES+=("$dockerfile")
  fi
done < <(find "$REPO_ROOT" -name "Dockerfile*" -not -path "*/vendor/*" -print0)

# --- 5. Update golang image references in YAML/YML files (e.g., docker-compose.yml) ---
# Matches patterns like:
#   image: golang:1.25.7-bullseye
#   image: "golang:1.25.7"
# Skips lines using variables like ${GO_VERSION} (already handled in step 1).
while IFS= read -r -d '' yamlfile; do
  if grep -qE 'golang:[0-9]+\.[0-9]+\.[0-9]+' "$yamlfile"; then
    "${SED_INPLACE[@]}" -E "s/(golang:)[0-9]+\.[0-9]+\.[0-9]+/\1${NEW_VERSION}/g" "$yamlfile"
    UPDATED_FILES+=("$yamlfile")
  fi
done < <(find "$REPO_ROOT" \( -name "*.yml" -o -name "*.yaml" \) -not -path "*/vendor/*" -print0)

# --- Summary ---
echo ""
echo "========================================="
echo " Go version updated to: ${NEW_VERSION}"
echo "========================================="
echo ""

if [[ ${#UPDATED_FILES[@]} -eq 0 ]]; then
  echo "No files were updated."
else
  # Deduplicate the list
  printf '%s\n' "${UPDATED_FILES[@]}" | sort -u | while IFS= read -r f; do
    echo "  ✓ ${f#"$REPO_ROOT/"}"
  done
fi

echo ""
echo "Done. Please review the changes and run 'go mod tidy' in each module directory."

