#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl stop splunk-otel-collector.service
    systemctl disable splunk-otel-collector.service
fi
