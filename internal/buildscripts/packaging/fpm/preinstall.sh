#!/bin/sh

getent passwd splunk-otel-collector >/dev/null || \
    useradd --system --user-group --no-create-home --shell /sbin/nologin splunk-otel-collector
