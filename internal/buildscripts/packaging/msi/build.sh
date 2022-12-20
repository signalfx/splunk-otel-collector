#!/bin/bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
#REPO_DIR="${"$(git rev-parse --show-toplevel || true)":-"$GITHUB_WORKSPACE"}"
REPO_DIR="$GITHUB_WORKSPACE"
[ "$(realpath"$SCRIPT_DIR/../../../../")" = "$REPO_DIR" ] || exit 1
cd "$REPO_DIR"
SMART_AGENT_RELEASE_PATH="${SCRIPT_DIR}/../smart-agent-release.txt"

VERSION="${1:-}"
SMART_AGENT_RELEASE="${2:-}"

get_version() {
    explicit_commit_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true )"
    heuristic_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true )"
    # Fail if not empty
    [ -z "$explicit_commit_tag" ] && [ -z "$heuristic_tag" ] && exit 1
    # Otherwise, return latest tag
    # Could probably default to $GITHUB_REF_NAME if not set...?
    latest_tag="${explicit_commit_tag:-"$heuristic_tag.1"}"
    echo "$latest_tag"
}

if [ -z "$SMART_AGENT_RELEASE" ]; then
    SMART_AGENT_RELEASE=$(cat "$SMART_AGENT_RELEASE_PATH")
fi

if [ -z "$VERSION" ]; then
    VERSION="$( get_version )"
fi

docker build -t msi-builder -f "${SCRIPT_DIR}/msi-builder/Dockerfile" "$REPO_DIR"
docker rm -fv msi-builder 2>/dev/null || true
docker run -d --name msi-builder msi-builder sleep inf
docker exec \
    -e SMART_AGENT_RELEASE="${SMART_AGENT_RELEASE}" \
    -e OUTPUT_DIR=/project/dist \
    -e VERSION="${VERSION#v}" \
    msi-builder /docker-entrypoint.sh
mkdir -p "${REPO_DIR}/dist"
docker cp msi-builder:/project/dist/splunk-otel-collector-${VERSION#v}-amd64.msi "${REPO_DIR}/dist/"
docker rm -fv msi-builder
