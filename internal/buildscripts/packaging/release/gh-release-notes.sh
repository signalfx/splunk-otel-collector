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
REPO_DIR="$( cd "$SCRIPT_DIR"/../../../../ && pwd )"

VERSION="$1"
LINUX_DIGEST="${2:-${REPO_DIR}/dist/linux_digest.txt}"
WINDOWS_DIGEST="${3:-${REPO_DIR}/dist/windows_digest.txt}"
CHANGELOG="${4:-${REPO_DIR}/CHANGELOG.md}"

changes="$( awk -v version="$VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
    echo "Failed to get changes for $VERSION from $CHANGELOG" >&2
    exit 1
fi

linux_digest="$( get_digest "$LINUX_DIGEST" )"

windows_digest="$( get_digest "$WINDOWS_DIGEST" )"

changes="""$changes

> Docker Images:
> - \`quay.io/signalfx/splunk-otel-collector:${VERSION#v}\` (digest: \`$linux_digest\`)
> - \`quay.io/signalfx/splunk-otel-collector-windows:${VERSION#v}\` (digest: \`$windows_digest\`)
"""

echo "$changes"
