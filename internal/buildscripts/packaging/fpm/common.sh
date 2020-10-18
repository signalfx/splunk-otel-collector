#!/bin/bash

FPM_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
REPO_DIR="$( cd "$FPM_DIR/../../../../" && pwd )"

PKG_NAME="splunk-otel-collector"
PKG_VENDOR="Splunk, Inc."
PKG_MAINTAINER="Splunk, Inc."
PKG_DESCRIPTION="Splunk OpenTelemetry Collector"
PKG_LICENSE="Apache 2.0"
PKG_URL="https://github.com/signalfx/splunk-otel-collector"

SERVICE_NAME="splunk-otel-collector"
SERVICE_USER="splunk-otel-collector"
SERVICE_GROUP="splunk-otel-collector"
PROCESS_NAME="otelcol"

SERVICE_PATH="$FPM_DIR/$SERVICE_NAME.service"
PREINSTALL_PATH="$FPM_DIR/preinstall.sh"
POSTINSTALL_PATH="$FPM_DIR/postinstall.sh"
PREUNINSTALL_PATH="$FPM_DIR/preuninstall.sh"

OTELCOL_INSTALL_PATH="/usr/bin/otelcol"
SPLUNK_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/splunk_config_linux.yaml"
SPLUNK_CONFIG_INSTALL_PATH="/etc/otel/collector/splunk_config_linux.yaml"
OTLP_CONFIG_REPO_PATH="$REPO_DIR/cmd/otelcol/config/collector/otlp_config_linux.yaml"
OTLP_CONFIG_INSTALL_PATH="/etc/otel/collector/otlp_config_linux.yaml"
SPLUNK_ENV_REPO_PATH="$FPM_DIR/splunk_env.example"
SPLUNK_ENV_INSTALL_PATH="/etc/otel/collector/splunk_env.example"


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


docker_cp() {
    local container="$1"
    local src="$2"
    local dest="$3"
    local dest_dir="$( dirname "$dest" )"

    echo "Copying $src to $container:$dest ..."
    docker exec $container mkdir -p "$dest_dir"
    docker cp "$src" $container:"$dest"
}


install_pkg() {
    local container="$1"
    local pkg_path="$2"
    local pkg_base=$( basename "$pkg_path" )

    echo "Installing $pkg_base ..."
    docker_cp $container "$pkg_path" /tmp/$pkg_base
    if [[ "${pkg_base##*.}" = "deb" ]]; then
        docker exec $container dpkg -i /tmp/$pkg_base
    else
        docker exec $container rpm -Uvh --replacepkgs /tmp/$pkg_base
    fi
}


uninstall_pkg() {
    local container="$1"
    local pkg_type="$2"
    local pkg_name="${3:-"$PKG_NAME"}"

    echo "Uninstalling $pkg_name ..."
    if [[ "$pkg_type" = "deb" ]]; then
        docker exec $container dpkg -r $pkg_name
    else
        docker exec $container rpm -e $pkg_name
    fi
}
