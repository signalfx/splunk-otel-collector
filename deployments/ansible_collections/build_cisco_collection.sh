#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="${SCRIPT_DIR}/signalfx"
DEST_DIR="${SCRIPT_DIR}/cisco"
COLLECTION_DIR="${SCRIPT_DIR}/dist"

if [[ ! -d "${SRC_DIR}" ]]; then
  echo "Source directory '${SRC_DIR}' does not exist." >&2
  exit 1
fi

rm -rf "${DEST_DIR}"
mkdir -p "${DEST_DIR}"

cp -R "${SRC_DIR}/." "${DEST_DIR}/"

README_MD="${DEST_DIR}/splunk_otel_collector/README.md"
if [[ ! -f "${README_MD}" ]]; then
  echo "Expected README not found at '${README_MD}'." >&2
  exit 1
fi

CHANGELOG_MD="${DEST_DIR}/splunk_otel_collector/CHANGELOG.md"
if [[ ! -f "${CHANGELOG_MD}" ]]; then
  echo "Expected CHANGELOG.md not found at '${CHANGELOG_MD}'." >&2
  exit 1
fi

GALAXY_YML="${DEST_DIR}/splunk_otel_collector/galaxy.yml"
if [[ ! -f "${GALAXY_YML}" ]]; then
  echo "Expected galaxy.yml not found at '${GALAXY_YML}'." >&2
  exit 1
fi

# Need to handle GNU and BSD sed.
if sed --version >/dev/null 2>&1; then
  sed -i -e 's/signalfx\.splunk_otel_collector/cisco.splunk_otel_collector/g' "${README_MD}"
  sed -i -e 's/^namespace:[[:space:]]*.*/namespace: cisco/' "${GALAXY_YML}"
  sed -i '/## ansible-v0.34.0/,$d' CHANGELOG_MD
else
  sed -i '' -e 's/signalfx\.splunk_otel_collector/cisco.splunk_otel_collector/g' "${README_MD}"
  sed -i '' -e 's/^namespace:[[:space:]]*.*/namespace: cisco/' "${GALAXY_YML}"
  sed -i '' '/## ansible-v0.34.0/,$d' CHANGELOG_MD
fi

mkdir -p "${COLLECTION_DIR}"
ansible-galaxy collection build "${DEST_DIR}"/splunk_otel_collector/ --output-path "${COLLECTION_DIR}"

rm -rf "${DEST_DIR}"
