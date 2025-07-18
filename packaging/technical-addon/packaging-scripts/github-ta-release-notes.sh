#!/bin/bash -eux

CHANGELOG="${REPO_DIR}/CHANGELOG.md"

changes="$( awk -v version="$VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
  changes="Release notes in progress."
fi

cat <<EOH
$changes

EOH
