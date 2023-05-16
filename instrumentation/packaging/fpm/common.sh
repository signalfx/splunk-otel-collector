#!/bin/bash

# Copyright Splunk Inc.
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

FPM_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$FPM_DIR/../../../" && pwd )"

PKG_VENDOR="Splunk, Inc."
PKG_MAINTAINER="Splunk, Inc."
PKG_DESCRIPTION="Splunk OpenTelemetry Auto Instrumentation"
PKG_LICENSE="Apache 2.0"
PKG_URL="https://github.com/signalfx/splunk-otel-collector"

JAVA_AGENT_RELEASE_PATH="${FPM_DIR}/../java-agent-release.txt"
JAVA_AGENT_RELEASE_URL="https://github.com/signalfx/splunk-otel-java/releases/"
JAVA_AGENT_INSTALL_PATH="/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"

get_version() {
    commit_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true )"
    if [[ -z "$commit_tag" ]]; then
        latest_tag="$( git -C "$REPO_DIR" describe --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true )"
        if [[ -n "$latest_tag" ]]; then
            echo "${latest_tag}-post"
        else
            echo "0.0.1-post"
        fi
    else
        echo "$commit_tag"
    fi
}

download_java_agent() {
    local tag="$1"
    local dest="$2"
    local dl_url=""
    if [[ "$tag" = "latest" ]]; then
      dl_url="$JAVA_AGENT_RELEASE_URL/latest/download/splunk-otel-javaagent.jar"
    else
      dl_url="$JAVA_AGENT_RELEASE_URL/download/$tag/splunk-otel-javaagent.jar"
    fi

    echo "Downloading $dl_url ..."
    mkdir -p "$( dirname $dest )"
    curl -sfL "$dl_url" -o "$dest"
}
