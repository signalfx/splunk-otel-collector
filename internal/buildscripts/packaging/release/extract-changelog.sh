#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR"/../../../../ && pwd )"

VERSION="$1"
CHANGELOG="${2:-${REPO_DIR}/CHANGELOG.md}"

changes="$( awk -v version="$VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
    echo "Failed to get changes for $VERSION from $CHANGELOG" >&2
    exit 1
fi

echo "$changes"
