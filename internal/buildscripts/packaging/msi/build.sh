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
FLUENTD_CONFIG="./internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd/fluent.conf"
FLUENTD_CONFD="./internal/buildscripts/packaging/msi/fluentd/conf.d"
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
    --otelcol PATH:     Relative path from the repo base directory to the otelcol exe.
                        Defaults to '$OTELCOL'.
    --config PATH:      Relative path from the repo base directory to the agent config.
                        Defaults to '$CONFIG'.
    --fluentd PATH:     Relative path from the repo base directory to the fluentd config.
                        Defaults to '$FLUENTD_CONFIG'.
    --eventlog PATH:    Relative path from the repo base directory to the eventlog config.
                        Defaults to '$EVENTLOG_CONFIG'.
    --output DIR:       Directory to save the MSI.
                        Defaults to '$OUTPUT_DIR'.

EOH
}

parse_args_and_build() {
    local otelcol="$OTELCOL"
    local config="$CONFIG"
    local fluentd_config="$FLUENTD_CONFIG"
    local fluentd_confd="$FLUENTD_CONFD"
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
            --fluentd)
                fluentd_config="$2"
                shift 1
                ;;
            --fluentd-confd)
                fluentd_confd="$2"
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

    docker_run="docker run --rm -v ${REPO_DIR}:/work -w /work $IMAGE_NAME"
    build_dir="${output}/build"
    files_dir="${build_dir}/msi"
    msi_name="splunk-otel-collector-${version}-amd64.msi"

    if [ -d "$build_dir" ]; then
        rm -rf "$build_dir"
    fi

    mkdir -p "${files_dir}/fluentd/conf.d"
    cp "$config" "${files_dir}/config.yaml"
    cp "$fluentd_config" "${files_dir}/fluentd/td-agent.conf"
    cp "${fluentd_confd}"/*.conf "${files_dir}/fluentd/conf.d/"

    $docker_run heat dir "$files_dir" -srd -sreg -gg -template fragment -cg ConfigFiles -dr INSTALLDIR -out "${build_dir}/configfiles.wsx"
    $docker_run candle -arch x64 -out "${build_dir}/configfiles.wixobj" "${build_dir}/configfiles.wsx"
    $docker_run candle -arch x64 -out "${build_dir}/splunk-otel-collector.wixobj" -dVersion="$version" -dOtelcol="$otelcol" "$WXS_PATH"
    $docker_run light -ext WixUtilExtension.dll -sval -out "${output}/${msi_name}" -b "${files_dir}" "${build_dir}/splunk-otel-collector.wixobj" "${build_dir}/configfiles.wixobj"
    echo "MSI saved to ${output}/${msi_name}"
}

parse_args_and_build $@
