#!/bin/sh

# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

if command -v systemctl >/dev/null 2>&1; then
    systemctl disable splunk-otel-script.service
    if systemctl status splunk-otel-script.service >/dev/null 2>&1; then
        echo "Stopping splunk-otel-script service"
        systemctl stop splunk-otel-script.service
    fi
fi
