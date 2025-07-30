#!/bin/bash -eux

[[ -z "${ADDONS_SOURCE_DIR}" ]] && echo "ADDONS_SOURCE_DIR must be set" && exit 1
[[ -z "${TA_NAME}" ]] && echo "TA_NAME must be set" && exit 1

CHANGELOG="${ADDONS_SOURCE_DIR}/${TA_NAME}/CHANGELOG.md"
TA_VERSION=$(sed -n '/^\[id\]/,/^\[/{s/^version = \(.*\)/\1/p}' "$ADDONS_SOURCE_DIR/Splunk_TA_otel/default/app.conf" | tr -d ' \t\n\r')
changes="$( awk -v version="$TA_VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
  changes="Release notes in progress."
fi

cat <<EOH
$changes

EOH
