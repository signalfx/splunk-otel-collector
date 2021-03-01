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

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd $SCRIPT_DIR/../../../../ && pwd )"

IMAGE_NAME="felfert/wix:latest"
WXS_PATH="./internal/buildscripts/packaging/msi/splunk-otel-collector.wxs"
OTELCOL="./bin/otelcol_windows_amd64.exe"
CONFIG="./cmd/otelcol/config/collector/agent_config.yaml"
OUTPUT_DIR="./dist"

usage() {
    cat <<EOH >&2
usage: ${BASH_SOURCE[0]} [OPTIONS] VERSION

Description:
    Build the MSI with the '$IMAGE_NAME' docker image.
    By default, the MSI is saved as '${OUTPUT_DIR}/splunk-otel-collector-VERSION-amd64.msi'.

Required Arguments:
    VERSION:        The version for the MSI. The version should be in the form "N.N.N" or "N.N.N.N".

OPTIONS:
    --otelcol PATH: Relative path from the repo base directory to the otelcol exe.
                    Defaults to '$OTELCOL'.
    --config PATH:  Relative path from the repo base directory to the agent config.
                    Defaults to '$CONFIG'.
    --output DIR:   Directory to save the MSI.
                    Defaults to '$OUTPUT_DIR'.

EOH
}

parse_args_and_build() {
    local otelcol="$OTELCOL"
    local config="$CONFIG"
    local output="$OUTPUT_DIR"
    local version=

    while [ -n "${1-}" ]; do
        case $1 in
            --otelcol)
                otelcol="$2"
                shift 1
                ;;
            --config)
                config="$2"
                shift 1
                ;;
            --output)
                output="$2"
                shift 1
                ;;
            -*)
                echo "Unknown option '$1'"
                echo
                usage
                exit 1
                ;;
            *)
                version="$1"
                ;;
        esac
        shift 1
    done

    if [[ -z "$version" ]]; then
        echo "Required VERSION argument not specified" >&2
        echo
        usage
        exit 1
    elif [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
        echo "Invalid version '$version'" >&2
        echo "The version should be in the form N.N.N or N.N.N.N" >&2
        exit 1
    fi

    docker run --rm -v "${REPO_DIR}":/work -w /work $IMAGE_NAME candle -arch x64 -dOtelcol="$otelcol" -dVersion="$version" -dConfig="$config" "$WXS_PATH"

    docker run --rm -v "${REPO_DIR}":/work -w /work $IMAGE_NAME light splunk-otel-collector.wixobj -sval

    mkdir -p "$output"

    mv -f splunk-otel-collector.msi "${output}/splunk-otel-collector-${version}-amd64.msi"

    echo "MSI saved to ${output}/splunk-otel-collector-${version}-amd64.msi"
}

parse_args_and_build $@
