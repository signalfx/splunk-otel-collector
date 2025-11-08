#!/bin/bash

set -euo pipefail

get_digest() {
    local digest_file="$1"
    local digest=""

    if [[ ! -f "$digest_file" ]]; then
        echo "$digest_file not found" >&2
        return 1
    fi

    digest="$( cat "$digest_file" | tr -dc '[[:print:]]' | sed 's|\[.*@\(sha256:.*\)\]|\1|' )"

    if [[ ! "$digest" =~ ^sha256:[A-Fa-f0-9]{64}$ ]]; then
        echo "Failed to get digest from $digest_file" >&2
        return 1
    fi

    echo "$digest"
}

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR"/../../ && pwd )"

VERSION="$1"
MULTIARCH_DIGEST="$( get_digest "${REPO_DIR}/dist/multiarch_digest.txt" )"
# WINDOWS_MULTIARCH_DIGEST="$( get_digest "${REPO_DIR}/dist/windows_multiarch_digest.txt" )"
CHANGELOG="${REPO_DIR}/CHANGELOG.md"

changes="$( awk -v version="$VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
  changes="Release notes in progress."
fi

cat <<EOH
$changes

> Docker Image Manifests:
> - Linux (amd64, arm64, ppc64le) and Windows (2019 amd64, 2022 amd64):
>   - \`quay.io/signalfx/splunk-otel-collector:${VERSION#v}\`
>   - digest: \`$MULTIARCH_DIGEST\`
EOH
