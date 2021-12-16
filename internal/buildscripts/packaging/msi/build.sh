#!/bin/bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/../../../../" && pwd )"
SMART_AGENT_RELEASE_PATH="${SCRIPT_DIR}/../smart-agent-release.txt"
JMX_METRIC_GATHERER_RELEASE_PATH="${SCRIPT_DIR}/../jmx-metric-gatherer-release.txt"

VERSION="${1:-}"
SMART_AGENT_RELEASE="${2:-}"
JMX_METRIC_GATHERER_RELEASE="${3:-}"

get_version() {
    commit_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true )"
    if [[ -z "$commit_tag" ]]; then
        latest_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true )"
        if [[ -n "$latest_tag" ]]; then
            echo "${latest_tag}.1"
        else
            echo "0.0.1"
        fi
    else
        echo "$commit_tag"
    fi
}

if [ -z "$SMART_AGENT_RELEASE" ]; then
    SMART_AGENT_RELEASE=$(cat "$SMART_AGENT_RELEASE_PATH")
fi

if [ -z "$JMX_METRIC_GATHERER_RELEASE" ]; then
    JMX_METRIC_GATHERER_RELEASE=$(cat "$JMX_METRIC_GATHERER_RELEASE_PATH")
fi

if [ -z "$VERSION" ]; then
    VERSION="$( get_version )"
fi

docker build -t msi-builder -f "${SCRIPT_DIR}/msi-builder/Dockerfile" "$REPO_DIR"
docker rm -fv msi-builder 2>/dev/null || true
docker run -d --name msi-builder msi-builder sleep inf
docker exec \
    -e SMART_AGENT_RELEASE="${SMART_AGENT_RELEASE}" \
    -e JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE}" \
    -e OUTPUT_DIR=/project/dist \
    -e VERSION="${VERSION#v}" \
    msi-builder /docker-entrypoint.sh
mkdir -p "${REPO_DIR}/dist"
docker cp msi-builder:/project/dist/splunk-otel-collector-${VERSION#v}-amd64.msi "${REPO_DIR}/dist/"
docker rm -fv msi-builder
