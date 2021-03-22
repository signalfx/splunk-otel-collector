#!/bin/sh

if [ -f /usr/lib/splunk-otel-collector/agent-bundle/bin/patch-interpreter ]; then
    /usr/lib/splunk-otel-collector/agent-bundle/bin/patch-interpreter /usr/lib/splunk-otel-collector/agent-bundle
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    systemctl enable splunk-otel-collector.service
fi
