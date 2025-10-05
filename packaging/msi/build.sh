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

if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" && "$OSTYPE" != "win32" ]]; then
    echo "Running on Non-Windows system"
    echo "This script should be run on Git Bash on a Windows box with WiX Toolset already installed"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/../../" && pwd )"
JMX_METRIC_GATHERER_RELEASE_PATH="${SCRIPT_DIR}/../jmx-metric-gatherer-release.txt"

VERSION="${1:-}"
JMX_METRIC_GATHERER_RELEASE="${2:-}"

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

# Verify WiX Toolset required version
expected_candle_version="Windows Installer XML Toolset Compiler version 3.14.0.8606"
if ! candle_first_line="$(candle.exe -? 2>/dev/null | head -n 1)"; then
    echo "Error: candle.exe not found or failed to run. Ensure WiX Toolset 3.14.0.8606 is installed and in PATH."
    echo "Latest version of 3.14 introduces an issue with elevation, see https://github.com/signalfx/splunk-otel-collector/pull/4688"
    exit 1
fi
if [[ "$candle_first_line" != "$expected_candle_version" ]]; then
    echo "Error: Unexpected candle.exe version."
    echo " Got:      '$candle_first_line'"
    echo " Expected: '$expected_candle_version'"
    exit 1
fi

if find "$REPO_DIR/packaging/msi" -name "*.wxs" -print0 | xargs -0 grep -q "RemoveFolderEx"; then
    echo "Custom action 'RemoveFolderEx' can't be used without corresponding WiX upgrade due to CVE-2024-29188."
    exit 1
fi

if ! test -f "$REPO_DIR/dist/agent-bundle_windows_amd64.zip"; then
    echo "$REPO_DIR/dist/agent-bundle_windows_amd64.zip not found! Either download a pre-built bundle to $REPO_DIR/dist/, or run '$REPO_DIR/packaging/bundle/scripts/windows/make.ps1 bundle' on a windows host and copy it to $REPO_DIR/dist/."
    exit 1
fi

OUTPUT_DIR="$REPO_DIR/dist" \
REPO_DIR="$REPO_DIR" \
WORK_DIR="$REPO_DIR/work" \
VERSION="$MSI_VERSION" \
JMX_METRIC_GATHERER_RELEASE="${JMX_METRIC_GATHERER_RELEASE}" \
    "$SCRIPT_DIR/msi-builder/build-launcher.sh"
