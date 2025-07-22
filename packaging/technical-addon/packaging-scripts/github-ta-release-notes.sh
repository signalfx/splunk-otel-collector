#!/bin/bash -eux

CHANGELOG="${ADDONS_SOURCE_DIR}/${TA_NAME}/CHANGELOG.md"

changes="$( awk -v version="$TA_VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
  changes="Release notes in progress."
fi

cat <<EOH
$changes

EOH
