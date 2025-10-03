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

MSI_SRC_DIR=${MSI_SRC_DIR:-"${REPO_DIR}/packaging/msi"}
WXS_PATH="${MSI_SRC_DIR}/splunk-otel-collector.wxs"
OTELCOL="${REPO_DIR}/bin/otelcol_windows_amd64.exe"
AGENT_CONFIG="${REPO_DIR}/cmd/otelcol/config/collector/agent_config.yaml"
GATEWAY_CONFIG="${REPO_DIR}/cmd/otelcol/config/collector/gateway_config.yaml"
FLUENTD_CONFIG=${FLUENTD_CONFIG:-"${REPO_DIR}/packaging/fpm/etc/otel/collector/fluentd/fluent.conf"}
FLUENTD_CONFD="${MSI_SRC_DIR}/fluentd/conf.d"
SUPPORT_BUNDLE_SCRIPT=${SUPPORT_BUNDLE_SCRIPT:-"${MSI_SRC_DIR}/splunk-support-bundle.ps1"}
SPLUNK_ICON="${MSI_SRC_DIR}/splunk.ico"
OUTPUT_DIR="${REPO_DIR}/dist"
JMX_METRIC_GATHERER_RELEASE="1.29.0"

usage() {
    cat <<EOH >&2
usage: ${BASH_SOURCE[0]} [OPTIONS] VERSION

Description:
    Build the Splunk OpenTelemetry MSI from the project available at ${PROJECT_DIR}.
    By default, the MSI is saved as '${OUTPUT_DIR}/splunk-otel-collector-VERSION-amd64.msi'.

OPTIONS:
    --otelcol PATH                    Absolute path to the otelcol exe.
                                      Defaults to '$OTELCOL'.
    --agent-config PATH               Absolute path to the agent config.
                                      Defaults to '$AGENT_CONFIG'.
    --gateway-config PATH             Absolute path to the gateway config.
                                      Defaults to '$GATEWAY_CONFIG'.
    --fluentd PATH                    DEPRECATED: Fluentd support has been deprecated and will be removed in a future release.
                                      Please refer to documentation for more information:
                                      https://github.com/signalfx/splunk-otel-collector/blob/main/docs/deprecations/fluentd-support.md
                                      Absolute path to the fluentd config.
                                      Defaults to '$FLUENTD_CONFIG'.
    --fluentd-confd PATH              DEPRECATED: Fluentd support has been deprecated and will be removed in a future release.
                                      Please refer to documentation for more information:
                                      https://github.com/signalfx/splunk-otel-collector/blob/main/docs/deprecations/fluentd-support.md
                                      Absolute path to the conf.d.
                                      Defaults to '$FLUENTD_CONFD'.
    --support-bundle PATH             Absolute path to the support bundle script.
                                      Defaults to '$SUPPORT_BUNDLE_SCRIPT'.
    --jmx-metric-gatherer VERSION     The released version of the JMX Metric Gatherer JAR to include (will be downloaded).
                                      Defaults to '$JMX_METRIC_GATHERER_RELEASE'.
    --splunk-icon PATH                Absolute path to the splunk.ico.
                                      Defaults to '$SPLUNK_ICON'.
    --output DIR                      Directory to save the MSI.
                                      Defaults to '$OUTPUT_DIR'.
    --skip-build-dir-removal          Skip removing the build directory before building the MSI.
EOH
}

parse_args_and_build() {
    local otelcol="$OTELCOL"
    local agent_config="$AGENT_CONFIG"
    local gateway_config="$GATEWAY_CONFIG"
    local fluentd_config="$FLUENTD_CONFIG"
    local fluentd_confd="$FLUENTD_CONFD"
    local support_bundle="$SUPPORT_BUNDLE_SCRIPT"
    local jmx_metric_gatherer_release="$JMX_METRIC_GATHERER_RELEASE"
    local splunk_icon="$SPLUNK_ICON"
    local output="$OUTPUT_DIR"
    local version=
    local skip_build_dir_removal=

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
                echo "[WARNING] DEPRECATED: Fluentd support has been deprecated and will be removed in a future release. Please refer to documentation for more information: https://github.com/signalfx/splunk-otel-collector/blob/main/docs/deprecations/fluentd-support.md"
                fluentd_config="$2"
                shift 1
                ;;
            --fluentd-confd)
                echo "[WARNING] DEPRECATED: Fluentd support has been deprecated and will be removed in a future release. Please refer to documentation for more information: https://github.com/signalfx/splunk-otel-collector/blob/main/docs/deprecations/fluentd-support.md"
                fluentd_confd="$2"
                shift 1
                ;;
            --support-bundle)
                support_bundle="$2"
                shift 1
                ;;
            --jmx-metric-gatherer)
                jmx_metric_gatherer_release="$2"
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
            --skip-build-dir-removal)
                skip_build_dir_removal=1
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
    build_dir="${WORK_DIR}/build"
    files_dir="${build_dir}/msi"
    msi_name="splunk-otel-collector-${version}-amd64.msi"

    if [ -z "$skip_build_dir_removal" ] && [ -d "$build_dir" ]; then
        rm -rf "$build_dir"
    else
        echo "Skipping build directory removal"
    fi

    mkdir -p "${files_dir}/fluentd/conf.d"
    cp "$support_bundle" "${files_dir}/splunk-support-bundle.ps1"
    cp "$agent_config" "${files_dir}/agent_config.yaml"
    cp "$gateway_config" "${files_dir}/gateway_config.yaml"
    cp "$fluentd_config" "${files_dir}/fluentd/td-agent.conf"
    cp "${fluentd_confd}"/*.conf "${files_dir}/fluentd/conf.d/"

    if [ -z "$skip_build_dir_removal" ]; then
        unzip -d "$files_dir" "${OUTPUT_DIR}/agent-bundle_windows_amd64.zip"
        rm -f "${OUTPUT_DIR}/agent-bundle_windows_amd64.zip"
    else
        echo "Skipping unzipping agent bundle"
    fi

    jmx_metrics_jar="${build_dir}/opentelemetry-java-contrib-jmx-metrics.jar"
    if [ -f "${jmx_metrics_jar}" ]; then
        echo "JMX Metric Gatherer already downloaded"
    else
        download_jmx_metric_gatherer "$jmx_metric_gatherer_release" "$build_dir"
    fi

    cd ${WORK_DIR}

    configFilesWsx="${build_dir}/configfiles.wsx"
    heat dir "$files_dir" -srd -sreg -gg -template fragment -cg ConfigFiles -dr INSTALLDIR -out "${configFilesWsx}"

    configFilesWixObj="${build_dir}/configfiles.wixobj"
    candle -arch x64 -out "${configFilesWixObj}" "${configFilesWsx}"

    collectorWixObj="${build_dir}/splunk-otel-collector.wixobj"
    candle -arch x64 -out "${collectorWixObj}" -dVersion="$version" -dOtelcol="$otelcol" -dJmxMetricsJar="$jmx_metrics_jar" "${WXS_PATH}"

    msi="${build_dir}/${msi_name}"
    light -ext WixUtilExtension.dll -sval -out "${msi}" -b "${files_dir}" "${collectorWixObj}" "${configFilesWixObj}"

    mkdir -p $output
    cp "${msi}" "${output}/${msi_name}"
    { set +x; } 2>/dev/null

    echo "MSI saved to ${output}/${msi_name}"
}

download_jmx_metric_gatherer() {
    local version="$1"
    local build_dir="$2"
    jmx_filename="opentelemetry-java-contrib-jmx-metrics.jar"
    JMX_METRIC_GATHERER_RELEASE_DL_URL="https://github.com/open-telemetry/opentelemetry-java-contrib/releases/download/$version/opentelemetry-jmx-metrics.jar"
    echo "Downloading ${JMX_METRIC_GATHERER_RELEASE_DL_URL}"

    mkdir -p "${build_dir}"
    curl -sL "$JMX_METRIC_GATHERER_RELEASE_DL_URL" -o "${build_dir}/${jmx_filename}"
}

parse_args_and_build $@
