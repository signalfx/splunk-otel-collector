#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl enable splunk-otel-collector.service
    if [ -f /etc/splunk-otel-collector/config.yaml ]; then
        systemctl start splunk-otel-collector.service
    fi
fi
