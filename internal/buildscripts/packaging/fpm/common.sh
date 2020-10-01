#!/bin/bash

FPM_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"

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
        docker exec $container rpm -ivh /tmp/$pkg_base
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
