#!/bin/sh

# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

getent passwd splunk-otel-collector >/dev/null || \
    useradd --system --user-group --home-dir /etc/otel/collector --no-create-home --shell $(command -v nologin) splunk-otel-collector

if command -v systemctl >/dev/null 2>&1 && systemctl status splunk-otel-script.service >/dev/null 2>&1; then
    echo "Stopping splunk-otel-script service"
    systemctl stop splunk-otel-script.service
    systemctl disable splunk-otel-script.service
fi
