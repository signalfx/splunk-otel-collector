#!/bin/sh

set -e

# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

echo "Installing Python library requirements..."
pip3 install -r /usr/lib/splunk-otel-script/requirements.txt

if [ -d /usr/lib/splunk-otel-script ]; then
    chown -R splunk-otel-collector:splunk-otel-collector /usr/lib/splunk-otel-script
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    systemctl enable splunk-otel-script.service
    echo "Starting splunk-otel-script service"
    ls -lah /usr/lib/splunk-otel-script/
    systemctl restart splunk-otel-script.service
fi
