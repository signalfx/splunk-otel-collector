#!/bin/sh

getent passwd splunk-otel-collector >/dev/null || \
    useradd --system --user-group --home-dir /etc/otel/collector --no-create-home --shell $(command -v nologin) splunk-otel-collector

if command -v systemctl >/dev/null 2>&1 && systemctl status splunk-otel-collector.service >/dev/null 2>&1; then
    echo "Stopping splunk-otel-collector service"
    systemctl stop splunk-otel-collector.service
    systemctl disable splunk-otel-collector.service
fi
