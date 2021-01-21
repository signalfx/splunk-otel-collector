#!/bin/sh

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    systemctl enable splunk-otel-collector.service
fi
