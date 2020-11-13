#!/bin/sh

if command -v systemctl >/dev/null 2>&1 && [ -f /usr/lib/systemd/system/splunk-otel-collector.service ]; then
    systemctl daemon-reload
    systemctl enable splunk-otel-collector.service
fi
