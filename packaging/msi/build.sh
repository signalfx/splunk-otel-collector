#!/bin/bash

# Copyright 2020 Splunk, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/../../" && pwd )"
JMX_METRIC_GATHERER_RELEASE_PATH="${SCRIPT_DIR}/../jmx-metric-gatherer-release.txt"

VERSION="${1:-}"
DOCKER_REPO="${2:-docker.io}"
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

if [ -z "$JMX_METRIC_GATHERER_RELEASE" ]; then
    JMX_METRIC_GATHERER_RELEASE=$(cat "$JMX_METRIC_GATHERER_RELEASE_PATH")
fi

if [ -z "$VERSION" ]; then
    VERSION="$( get_version )"
fi

# Convert pre-release version format for MSI compatibility
# e.g., v0.130.1-rc.0 -> 0.130.1.0, v0.130.1-beta.1 -> 0.130.1.1
convert_version_for_msi() {
    local version="$1"
    version="${version#v}"
    
    if [[ "$version" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-(rc|beta)\.([0-9]+)$ ]]; then
        local base_version="${BASH_REMATCH[1]}"
        local prerelease_type="${BASH_REMATCH[2]}"
        local prerelease_number="${BASH_REMATCH[3]}"
        
        if [[ "$prerelease_type" == "rc" ]]; then
            echo "${base_version}.$prerelease_number"
        elif [[ "$prerelease_type" == "beta" ]]; then
            echo "${base_version}.$((100 + prerelease_number))"
        fi
    else
        echo "$version"
    fi
}

MSI_VERSION=$(convert_version_for_msi "$VERSION")

mkdir -p "${REPO_DIR}/dist"

# On Windows directly invoke Wix tools, on other platforms go with the docker commands
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    echo "Running on Windows"
    OUTPUT_DIR="${REPO_DIR}/dist/" VERSION="${MSI_VERSION}" JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE}" "${SCRIPT_DIR}/msi-builder/docker-entrypoint.sh"
else
    echo "Running on Unix-like system"

    docker build -t msi-builder \
        --build-arg JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE}" \
        --build-arg DOCKER_REPO="$DOCKER_REPO" \
        -f "${SCRIPT_DIR}/msi-builder/Dockerfile" \
        "$REPO_DIR"
    docker rm -fv msi-builder 2>/dev/null || true
    docker run -d --name msi-builder msi-builder sleep inf
    docker exec \
        -e OUTPUT_DIR=/project/dist \
        -e VERSION="$MSI_VERSION" \
        msi-builder /docker-entrypoint.sh
    docker cp msi-builder:/project/dist/splunk-otel-collector-${MSI_VERSION}-amd64.msi "${REPO_DIR}/dist/"
    docker rm -fv msi-builder
fi
