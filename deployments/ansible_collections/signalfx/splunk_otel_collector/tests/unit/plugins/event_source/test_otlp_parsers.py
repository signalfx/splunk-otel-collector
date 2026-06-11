"""Tests for OTLP payload parsers."""

from __future__ import annotations

import pytest

from extensions.eda.plugins.event_source._otlp_parsers import (
    extract_resource_attributes,
    nanos_to_iso,
    parse_logs_json,
)
from .conftest import SAMPLE_LOG_JSON


def test_extract_resource_attributes():
    """Test basic string attribute extraction."""
    attrs = [
        {"key": "host.name", "value": {"stringValue": "web-01"}},
        {"key": "service.name", "value": {"stringValue": "nginx"}},
    ]
    result = extract_resource_attributes(attrs)
    assert result == {
        "host.name": "web-01",
        "service.name": "nginx",
    }


def test_extract_resource_attributes_int_value():
    """Test integer value extraction."""
    attrs = [
        {"key": "host.name", "value": {"stringValue": "web-01"}},
        {"key": "port", "value": {"intValue": "8080"}},
    ]
    result = extract_resource_attributes(attrs)
    assert result == {
        "host.name": "web-01",
        "port": 8080,
    }


def test_extract_resource_attributes_bool_value():
    """Test boolean value extraction."""
    attrs = [
        {"key": "enabled", "value": {"boolValue": True}},
        {"key": "debug", "value": {"boolValue": False}},
    ]
    result = extract_resource_attributes(attrs)
    assert result == {
        "enabled": True,
        "debug": False,
    }


def test_nanos_to_iso():
    """Test nanosecond timestamp conversion to ISO 8601."""
    # 2024-06-11T10:10:00Z
    result = nanos_to_iso("1718100600000000000")
    assert result == "2024-06-11T10:10:00Z"


def test_nanos_to_iso_zero():
    """Test zero timestamp returns empty string."""
    result = nanos_to_iso("0")
    assert result == ""


def test_parse_logs_json_single_record():
    """Test parsing a single log record from OTLP JSON."""
    result = parse_logs_json(SAMPLE_LOG_JSON)

    assert len(result) == 1
    event = result[0]

    # Check signal type
    assert event["signal_type"] == "log"

    # Check resource attributes
    assert event["resource"]["host.name"] == "web-01"
    assert event["resource"]["service.name"] == "nginx"
    assert event["resource"]["service.version"] == "1.25"
    assert event["resource"]["deployment.environment.name"] == "production"

    # Check log data
    assert event["log"]["severity_text"] == "ERROR"
    assert event["log"]["severity_number"] == 17
    assert event["log"]["body"] == "Connection refused to upstream"
    assert event["log"]["timestamp"] == "2024-06-11T10:10:00Z"
    assert event["log"]["attributes"]["error.type"] == "ConnectionError"

    # Check metadata
    assert event["meta"]["endpoint"] == "v1/logs"
    assert event["meta"]["hosts"] == "web-01"


def test_parse_logs_json_missing_host_name():
    """Test fallback to localhost when host.name is missing."""
    payload = {
        "resourceLogs": [
            {
                "resource": {
                    "attributes": [
                        {"key": "service.name", "value": {"stringValue": "test"}},
                    ]
                },
                "scopeLogs": [
                    {
                        "logRecords": [
                            {
                                "timeUnixNano": "1718100600000000000",
                                "severityNumber": 9,
                                "severityText": "INFO",
                                "body": {"stringValue": "Test log"},
                                "attributes": [],
                            }
                        ]
                    }
                ],
            }
        ]
    }

    result = parse_logs_json(payload)
    assert len(result) == 1
    assert result[0]["meta"]["hosts"] == "localhost"


def test_parse_logs_json_empty_payload():
    """Test parsing empty payload returns empty list."""
    payload = {"resourceLogs": []}
    result = parse_logs_json(payload)
    assert result == []
