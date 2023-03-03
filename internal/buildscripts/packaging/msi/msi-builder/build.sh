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
TRANSLATESFX="/project/bin/translatesfx_windows_amd64.exe"
AGENT_CONFIG="/project/cmd/otelcol/config/collector/agent_config.yaml"
GATEWAY_CONFIG="/project/cmd/otelcol/config/collector/gateway_config.yaml"
FLUENTD_CONFIG="/project/internal/buildscripts/packaging/fpm/etc/otel/collector/fluentd/fluent.conf"
FLUENTD_CONFD="/project/internal/buildscripts/packaging/msi/fluentd/conf.d"
SUPPORT_BUNDLE_SCRIPT="/project/internal/buildscripts/packaging/msi/splunk-support-bundle.ps1"
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
    --translatesfx PATH:     Absolute path to the translatesfx exe.
                             Defaults to '$TRANSLATESFX'.
    --agent-config PATH:     Absolute path to the agent config.
                             Defaults to '$AGENT_CONFIG'.
    --gateway-config PATH:   Absolute path to the gateway config.
                             Defaults to '$GATEWAY_CONFIG'.
    --fluentd PATH:          Absolute path to the fluentd config.
                             Defaults to '$FLUENTD_CONFIG'.
    --fluentd-confd PATH:    Absolute path to the conf.d.
                             Defaults to '$FLUENTD_CONFD'.
    --support-bundle PATH:   Absolute path to the support bundle script.
                             Defaults to '$SUPPORT_BUNDLE_SCRIPT'.
    --splunk-icon PATH:      Absolute path to the splunk.ico.
                             Defaults to '$SPLUNK_ICON'.
    --output DIR:            Directory to save the MSI.
                             Defaults to '$OUTPUT_DIR'.

EOH
}

parse_args_and_build() {
    local otelcol="$OTELCOL"
    local translatesfx="$TRANSLATESFX"
    local agent_config="$AGENT_CONFIG"
    local gateway_config="$GATEWAY_CONFIG"
    local fluentd_config="$FLUENTD_CONFIG"
    local fluentd_confd="$FLUENTD_CONFD"
    local support_bundle="$SUPPORT_BUNDLE_SCRIPT"
    local splunk_icon="$SPLUNK_ICON"
    local output="$OUTPUT_DIR"
    local version=

    while [ -n "${1-}" ]; do
        case $1 in
            --otelcol)
                otelcol="$2"
                shift 1
                ;;
            --translatesfx)
                translatesfx="$2"
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
            --support-bundle)
                support_bundle="$2"
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
    cp "$support_bundle" "${files_dir}/splunk-support-bundle.ps1"
    cp "$agent_config" "${files_dir}/agent_config.yaml"
    cp "$gateway_config" "${files_dir}/gateway_config.yaml"
    cp "$fluentd_config" "${files_dir}/fluentd/td-agent.conf"
    cp "${fluentd_confd}"/*.conf "${files_dir}/fluentd/conf.d/"

    unzip -d "$files_dir" "${OUTPUT_DIR}/agent-bundle_windows_amd64.zip"
    rm -f "${OUTPUT_DIR}/agent-bundle_windows_amd64.zip"

    # kludge to satisfy relative path in splunk-otel-collector.wxs
    mkdir -p /work/internal/buildscripts/packaging/msi
    cp "${splunk_icon}" "/work/internal/buildscripts/packaging/msi/splunk.ico"

    cd /work
    configFilesWsx="${build_dir}/configfiles.wsx"
    heat dir "$files_dir" -srd -sreg -gg -template fragment -cg ConfigFiles -dr INSTALLDIR -out "${configFilesWsx//\//\\}"

    configFilesWixObj="${build_dir}/configfiles.wixobj"
    candle -arch x64 -out "${configFilesWixObj//\//\\}" "${configFilesWsx//\//\\}"

    collectorWixObj="${build_dir}/splunk-otel-collector.wixobj"
    candle -arch x64 -out "${collectorWixObj//\//\\}" -dVersion="$version" -dOtelcol="$otelcol" -dTranslatesfx="$translatesfx" "${WXS_PATH//\//\\}"

    msi="${build_dir}/${msi_name}"
    light -ext WixUtilExtension.dll -sval -out "${msi//\//\\}" -b "${files_dir//\//\\}" "${collectorWixObj//\//\\}" "${configFilesWixObj//\//\\}"

    mkdir -p $output
    cp "${msi}" "${output}/${msi_name}"
    { set +x; } 2>/dev/null

    echo "MSI saved to ${output}/${msi_name}"
}

parse_args_and_build $@
