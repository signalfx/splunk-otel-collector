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

PKG_NAME="splunk-otel-auto-instrumentation"
PKG_VENDOR="Splunk, Inc."
PKG_MAINTAINER="Splunk, Inc."
PKG_DESCRIPTION="Splunk OpenTelemetry Auto Instrumentation"
PKG_LICENSE="Apache 2.0"
PKG_URL="https://github.com/signalfx/splunk-otel-collector"

INSTALL_DIR="/usr/lib/splunk-instrumentation"
LIBSPLUNK_INSTALL_PATH="${INSTALL_DIR}/libsplunk.so"
JAVA_AGENT_INSTALL_PATH="${INSTALL_DIR}/splunk-otel-javaagent.jar"
CONFIG_DIR_REPO_PATH="${FPM_DIR}/etc/splunk/zeroconfig"
CONFIG_DIR_INSTALL_PATH="/etc/splunk/zeroconfig"
EXAMPLES_INSTALL_DIR="${INSTALL_DIR}/examples"
EXAMPLES_DIR="${FPM_DIR}/examples"

JAVA_AGENT_RELEASE_PATH="${FPM_DIR}/../java-agent-release.txt"
JAVA_AGENT_RELEASE_URL="https://github.com/signalfx/splunk-otel-java/releases/"
JAVA_AGENT_INSTALL_PATH="${INSTALL_DIR}/splunk-otel-javaagent.jar"

NODEJS_AGENT_RELEASE_PATH="${FPM_DIR}/../nodejs-agent-release.txt"
NODEJS_AGENT_RELEASE_URL="https://github.com/signalfx/splunk-otel-js/releases/"
NODEJS_AGENT_INSTALL_PATH="${INSTALL_DIR}/splunk-otel-js.tgz"

PREUNINSTALL_PATH="$FPM_DIR/preuninstall.sh"

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

download_nodejs_agent() {
    local tag="$1"
    local dest="$2"
    local dl_url="$NODEJS_AGENT_RELEASE_URL/download/$tag/splunk-otel-${tag#v}.tgz"

    echo "Downloading $dl_url ..."
    mkdir -p "$( dirname $dest )"
    curl -sfL "$dl_url" -o "$dest"
}

setup_files_and_permissions() {
    local arch="$1"
    local buildroot="$2"
    local libsplunk="$REPO_DIR/instrumentation/dist/libsplunk_${arch}.so"
    local java_agent_release="$(cat "$JAVA_AGENT_RELEASE_PATH")"
    local nodejs_agent_release="$(cat "$NODEJS_AGENT_RELEASE_PATH")"

    mkdir -p "$buildroot/$(dirname $LIBSPLUNK_INSTALL_PATH)"
    cp -f "$libsplunk" "$buildroot/$LIBSPLUNK_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$LIBSPLUNK_INSTALL_PATH"

    download_java_agent "$java_agent_release" "${buildroot}/${JAVA_AGENT_INSTALL_PATH}"
    sudo chmod 755 "$buildroot/$JAVA_AGENT_INSTALL_PATH"

    download_nodejs_agent "$nodejs_agent_release" "${buildroot}/${NODEJS_AGENT_INSTALL_PATH}"
    sudo chmod 644 "$buildroot/$NODEJS_AGENT_INSTALL_PATH"

    mkdir -p  "$buildroot/$CONFIG_DIR_INSTALL_PATH"
    cp -rf "$CONFIG_DIR_REPO_PATH"/* "$buildroot/$CONFIG_DIR_INSTALL_PATH"/
    sudo chmod -R 644 "$buildroot/$CONFIG_DIR_INSTALL_PATH"

    mkdir -p "$buildroot/$INSTALL_DIR"
    cp -rf "$EXAMPLES_DIR" "$buildroot/$INSTALL_DIR/"
    sudo chmod -R 644 "$buildroot/$EXAMPLES_INSTALL_DIR"

    sudo chown -R root:root "$buildroot"
}
