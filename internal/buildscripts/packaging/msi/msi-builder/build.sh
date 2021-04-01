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

WXS_PATH="/project/internal/buildscripts/packaging/msi/splunk-otel-collector.wxs"
OTELCOL="/project/bin/otelcol_windows_amd64.exe"
AGENT_CONFIG="/project/cmd/otelcol/config/collector/agent_config.yaml"
GATEWAY_CONFIG="/project/cmd/otelcol/config/collector/gateway_config.yaml"
FLUENTD_CONFIG="/project/internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd/fluent.conf"
FLUENTD_CONFD="/project/internal/buildscripts/packaging/msi/fluentd/conf.d"
SMART_AGENT_RELEASE="latest"
SPLUNK_ICON="/project/internal/buildscripts/packaging/msi/splunk.ico"
OUTPUT_DIR="/project/dist"

usage() {
    cat <<EOH >&2
usage: ${BASH_SOURCE[0]} [OPTIONS] VERSION

Description:
    Build the Splunk OpenTelemetry MSI from the project available at /project.
    By default, the MSI is saved as '${OUTPUT_DIR}/splunk-otel-collector-VERSION-amd64.msi'.

OPTIONS:
    --otelcol PATH:          Absolute path to the otelcol exe.
                             Defaults to '$OTELCOL'.
    --agent-config PATH:     Absolute path to the agent config.
                             Defaults to '$AGENT_CONFIG'.
    --gateway-config PATH:   Absolute path to the gateway config.
                             Defaults to '$GATEWAY_CONFIG'.
    --fluentd PATH:          Absolute path to the fluentd config.
                             Defaults to '$FLUENTD_CONFIG'.
    --fluentd-confd PATH:    Absolute path to the conf.d.
                             Defaults to '$FLUENTD_CONFD'.
    --smart-agent VERSION:   The released version of the Smart Agent bundle to include (will be downloaded).
                             Defaults to '$SMART_AGENT_RELEASE'.
    --splunk-icon PATH:      Absolute path to the splunk.ico.
                             Defaults to '$SPLUNK_ICON'.
    --output DIR:            Directory to save the MSI.
                             Defaults to '$OUTPUT_DIR'.

EOH
}

parse_args_and_build() {
    local otelcol="$OTELCOL"
    local agent_config="$AGENT_CONFIG"
    local gateway_config="$GATEWAY_CONFIG"
    local fluentd_config="$FLUENTD_CONFIG"
    local fluentd_confd="$FLUENTD_CONFD"
    local smart_agent_release="$SMART_AGENT_RELEASE"
    local splunk_icon="$SPLUNK_ICON"
    local output="$OUTPUT_DIR"
    local version=

    while [ -n "${1-}" ]; do
        case $1 in
            --otelcol)
                otelcol="$2"
                shift 1
                ;;
            --agent-config)
                agent_config="$2"
                shift 1
                ;;
            --gateway-config)
                gateway_config="$2"
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
            --smart-agent)
                smart_agent_release="$2"
                shift 1
                ;;
            --splunk-icon)
                splunk_icon="$2"
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

    set -x
    build_dir="/work/build"
    files_dir="${build_dir}/msi"
    msi_name="splunk-otel-collector-${version}-amd64.msi"

    if [ -d "$build_dir" ]; then
        rm -rf "$build_dir"
    fi

    mkdir -p "${files_dir}/fluentd/conf.d"
    cp "$agent_config" "${files_dir}/agent_config.yaml"
    cp "$gateway_config" "${files_dir}/gateway_config.yaml"
    cp "$fluentd_config" "${files_dir}/fluentd/td-agent.conf"
    cp "${fluentd_confd}"/*.conf "${files_dir}/fluentd/conf.d/"

    download_and_extract_smart_agent "$smart_agent_release" "$build_dir" "$files_dir"

    # kludge to satisfy relative path in splunk-otel-collector.wxs
    mkdir -p /work/internal/buildscripts/packaging/msi
    cp "${splunk_icon}" "/work/internal/buildscripts/packaging/msi/splunk.ico"

    cd /work
    configFilesWsx="${build_dir}/configfiles.wsx"
    heat dir "$files_dir" -srd -sreg -gg -template fragment -cg ConfigFiles -dr INSTALLDIR -out "${configFilesWsx//\//\\}"

    configFilesWixObj="${build_dir}/configfiles.wixobj"
    candle -arch x64 -out "${configFilesWixObj//\//\\}" "${configFilesWsx//\//\\}"

    collectorWixObj="${build_dir}/splunk-otel-collector.wixobj"
    candle -arch x64 -out "${collectorWixObj//\//\\}" -dVersion="$version" -dOtelcol="$otelcol" "${WXS_PATH//\//\\}"

    msi="${build_dir}/${msi_name}"
    light -ext WixUtilExtension.dll -sval -out "${msi//\//\\}" -b "${files_dir//\//\\}" "${collectorWixObj//\//\\}" "${configFilesWixObj//\//\\}"

    mkdir -p $output
    cp "${msi}" "${output}/${msi_name}"
    { set +x; } 2>/dev/null

    echo "MSI saved to ${output}/${msi_name}"
}

download_and_extract_smart_agent() {
    SMART_AGENT_RELEASE_URL="https://dl.signalfx.com/windows/release/zip"
    SMART_AGENT_LATEST_URL="${SMART_AGENT_RELEASE_URL}/latest/latest.txt"
    local version="$1"
    local build_dir="$2"
    local output_dir="$3/agent-bundle"

    if [ "$version" = "latest" ]; then
        version=$( curl -sL "$SMART_AGENT_LATEST_URL" )
        if [ -z "$version" ]; then
            echo "Failed to get version for latest release from ${SMART_AGENT_LATEST_URL}" >&2
            exit 1
        fi
    fi

    dl_url="$SMART_AGENT_RELEASE_URL/SignalFxAgent-$version-win64.zip"
    echo "Downloading ${dl_url}..."
    curl -sL "$dl_url" -o "${build_dir}/signalfx-agent.zip"

    unzip -d "$build_dir" "${build_dir}/signalfx-agent.zip"
    mv "${build_dir}/SignalFxAgent" "$output_dir"

    # Delete unnecessary files.
    rm -rf "${output_dir}/bin"
    rm -rf "${output_dir}/etc"
    find "$output_dir" -type d -name __pycache__ | xargs rm -rf {} \;
    find "$output_dir" -regextype sed -regex ".*py[co]" -delete
    # This defers resolving https://github.com/wixtoolset/issues/issues/5314, which appears to require building on windows.
    # Check if test file content's path is >= 126 (128 w/ 'Z:' prefix in wine).
    find "$output_dir" -wholename "*/tests/*" -exec bash -c 'if [ `echo "{}" | wc -c` -ge 126 ]; then rm -f {}; fi' \;
    rm -f "${build_dir}/signalfx-agent.zip"
}

parse_args_and_build $@
