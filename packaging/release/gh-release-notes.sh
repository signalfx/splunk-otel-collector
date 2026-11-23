#!/bin/bash

set -euo pipefail

get_digest() {
    local digest_file="$1"
    local repository="$2" record

    if [[ ! -f "$digest_file" ]]; then
        echo "$digest_file not found" >&2
        return 1
    fi

    record="$(< "$digest_file")"
    if [[ ! "$record" =~ ^\["$repository"@(sha256:[a-f0-9]{64})\]$ ]]; then
        echo "Failed to get digest from $digest_file" >&2
        return 1
    fi

    printf '%s\n' "${BASH_REMATCH[1]}"
}

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR"/../../ && pwd )"

VERSION="$1"
STAGING_TAG="$2"
QUAY_COLLECTOR_REPO="${QUAY_COLLECTOR_REPO:-quay.io/signalfx/splunk-otel-collector}"
DOCKER_HUB_COLLECTOR_REPO="${DOCKER_HUB_COLLECTOR_REPO:-docker.io/splunk/otel-collector}"
QUAY_DIGEST="$(get_digest "${REPO_DIR}/dist/docker/${QUAY_COLLECTOR_REPO}/multiarch_digest.txt" "$QUAY_COLLECTOR_REPO")"
DOCKER_HUB_DIGEST="$(get_digest "${REPO_DIR}/dist/docker/${DOCKER_HUB_COLLECTOR_REPO}/multiarch_digest.txt" "$DOCKER_HUB_COLLECTOR_REPO")"
PROMOTION_DIGEST="$(awk -v image="${QUAY_COLLECTOR_REPO}:${STAGING_TAG}" '$1 == image { print $2 }' \
  "${REPO_DIR}/dist/docker/${DOCKER_HUB_COLLECTOR_REPO}/images_multiarch.txt")"
if [[ "$QUAY_DIGEST" != "$PROMOTION_DIGEST" || "$DOCKER_HUB_DIGEST" != "$PROMOTION_DIGEST" ]]; then
  echo "Release-note digests must match the manifest selected for promotion" >&2
  exit 1
fi
CHANGELOG="${3:-${REPO_DIR}/CHANGELOG.md}"

changes="$( awk -v version="$VERSION" '/^## / { if (p) { exit }; if ($2 == version) { p=1; next } } p && NF' "$CHANGELOG" )"

if [[ -z "$changes" ]] || [[ "$changes" =~ ^[[:space:]]+$ ]]; then
  changes="Release notes in progress."
fi

cat <<EOH
$changes

> Docker Image Manifests:
> - Linux (amd64, arm64, ppc64le) and Windows (2019 amd64, 2022 amd64):
>   - Quay: \`${QUAY_COLLECTOR_REPO}:${VERSION#v}\` (digest: \`$QUAY_DIGEST\`)
>   - Docker Hub: \`${DOCKER_HUB_COLLECTOR_REPO}:${VERSION#v}\` (digest: \`$DOCKER_HUB_DIGEST\`)
EOH
