#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    systemctl enable splunk-otel-collector.service
    # only start the service after package is installed if the environment file exists
    if [ -f /etc/otel/collector/splunk_env ]; then
        systemctl start splunk-otel-collector.service
    fi
fi
