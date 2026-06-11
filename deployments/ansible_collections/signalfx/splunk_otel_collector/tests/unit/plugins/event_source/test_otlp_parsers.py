"""Tests for OTLP payload parsers."""

from __future__ import annotations

import json

import pytest
from google.protobuf.json_format import Parse
from opentelemetry.proto.collector.logs.v1.logs_service_pb2 import (
    ExportLogsServiceRequest,
)
from opentelemetry.proto.collector.metrics.v1.metrics_service_pb2 import (
    ExportMetricsServiceRequest,
)
from opentelemetry.proto.collector.trace.v1.trace_service_pb2 import (
    ExportTraceServiceRequest,
)

from extensions.eda.plugins.event_source._otlp_parsers import (
    extract_resource_attributes,
    nanos_to_iso,
    parse_logs_json,
    parse_logs_proto,
    parse_metrics_json,
    parse_metrics_proto,
    parse_traces_json,
    parse_traces_proto,
)
from .conftest import SAMPLE_LOG_JSON, SAMPLE_METRIC_JSON, SAMPLE_TRACE_JSON


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


def test_parse_metrics_json_histogram():
    """Test parsing a histogram metric from OTLP JSON."""
    result = parse_metrics_json(SAMPLE_METRIC_JSON)

    assert len(result) == 1
    event = result[0]

    # Check signal type
    assert event["signal_type"] == "metric"

    # Check resource attributes
    assert event["resource"]["host.name"] == "web-01"
    assert event["resource"]["service.name"] == "nginx"

    # Check metric data
    metric = event["metric"]
    assert metric["name"] == "http.server.request.duration"
    assert metric["type"] == "histogram"
    assert metric["unit"] == "s"

    # Check data points
    assert len(metric["data_points"]) == 1
    dp = metric["data_points"][0]
    assert dp["attributes"]["http.request.method"] == "GET"
    assert dp["start_time"] == "2024-06-11T10:09:00Z"
    assert dp["time"] == "2024-06-11T10:10:00Z"
    assert dp["count"] == 150
    assert dp["sum"] == 45.2
    assert dp["bucket_counts"] == [10, 50, 60, 25, 5]
    assert dp["explicit_bounds"] == [0.01, 0.05, 0.1, 0.5]

    # Check metadata
    assert event["meta"]["endpoint"] == "v1/metrics"
    assert event["meta"]["hosts"] == "web-01"


def test_parse_metrics_json_gauge():
    """Test parsing a gauge metric with asDouble value."""
    payload = {
        "resourceMetrics": [
            {
                "resource": {
                    "attributes": [
                        {"key": "host.name", "value": {"stringValue": "db-01"}},
                    ]
                },
                "scopeMetrics": [
                    {
                        "metrics": [
                            {
                                "name": "cpu.utilization",
                                "unit": "1",
                                "gauge": {
                                    "dataPoints": [
                                        {
                                            "attributes": [],
                                            "timeUnixNano": "1718100600000000000",
                                            "asDouble": 0.85,
                                        }
                                    ]
                                },
                            }
                        ]
                    }
                ],
            }
        ]
    }

    result = parse_metrics_json(payload)
    assert len(result) == 1

    metric = result[0]["metric"]
    assert metric["name"] == "cpu.utilization"
    assert metric["type"] == "gauge"
    assert metric["unit"] == "1"
    assert len(metric["data_points"]) == 1
    assert metric["data_points"][0]["value"] == 0.85
    assert metric["data_points"][0]["time"] == "2024-06-11T10:10:00Z"


def test_parse_metrics_json_sum():
    """Test parsing a sum metric with asInt value."""
    payload = {
        "resourceMetrics": [
            {
                "resource": {
                    "attributes": [
                        {"key": "host.name", "value": {"stringValue": "app-01"}},
                    ]
                },
                "scopeMetrics": [
                    {
                        "metrics": [
                            {
                                "name": "http.server.requests",
                                "unit": "1",
                                "sum": {
                                    "dataPoints": [
                                        {
                                            "attributes": [],
                                            "timeUnixNano": "1718100600000000000",
                                            "asInt": "5000",
                                            "isMonotonic": True,
                                        }
                                    ],
                                    "aggregationTemporality": 2,
                                },
                            }
                        ]
                    }
                ],
            }
        ]
    }

    result = parse_metrics_json(payload)
    assert len(result) == 1

    metric = result[0]["metric"]
    assert metric["name"] == "http.server.requests"
    assert metric["type"] == "sum"
    assert metric["unit"] == "1"
    assert len(metric["data_points"]) == 1
    assert metric["data_points"][0]["value"] == 5000
    assert metric["data_points"][0]["time"] == "2024-06-11T10:10:00Z"


def test_parse_metrics_json_empty():
    """Test parsing empty metrics payload returns empty list."""
    payload = {"resourceMetrics": []}
    result = parse_metrics_json(payload)
    assert result == []


def test_parse_traces_json_single_span():
    """Test parsing a single span from OTLP JSON."""
    result = parse_traces_json(SAMPLE_TRACE_JSON)

    assert len(result) == 1
    event = result[0]

    # Check signal type
    assert event["signal_type"] == "trace"

    # Check resource attributes
    assert event["resource"]["host.name"] == "web-01"
    assert event["resource"]["service.name"] == "api-gateway"

    # Check span data
    span = event["span"]
    assert span["trace_id"] == "abcdef1234567890abcdef1234567890"
    assert span["span_id"] == "1234567890abcdef"
    assert span["parent_span_id"] == ""
    assert span["name"] == "HTTP GET /api/users"
    assert span["kind"] == 2
    assert span["start_time"] == "2024-06-11T10:10:00Z"
    assert span["end_time"] == "2024-06-11T10:10:01Z"
    assert span["status_code"] == 2
    assert span["status_message"] == "Internal Server Error"
    assert span["attributes"]["http.response.status_code"] == 500
    assert span["attributes"]["http.request.method"] == "GET"

    # Check metadata
    assert event["meta"]["endpoint"] == "v1/traces"
    assert event["meta"]["hosts"] == "web-01"


def test_parse_traces_json_empty():
    """Test parsing empty traces payload returns empty list."""
    payload = {"resourceSpans": []}
    result = parse_traces_json(payload)
    assert result == []


def test_parse_logs_proto():
    """Test parsing a log record from OTLP protobuf."""
    # Create protobuf message from JSON
    proto_msg = Parse(json.dumps(SAMPLE_LOG_JSON), ExportLogsServiceRequest())
    proto_bytes = proto_msg.SerializeToString()

    # Parse protobuf bytes
    result = parse_logs_proto(proto_bytes)

    assert len(result) == 1
    event = result[0]

    # Check signal type
    assert event["signal_type"] == "log"

    # Check resource attributes
    assert event["resource"]["host.name"] == "web-01"

    # Check log data
    assert event["log"]["severity_text"] == "ERROR"
    assert event["log"]["severity_number"] == 17
    assert event["log"]["body"] == "Connection refused to upstream"

    # Check metadata
    assert event["meta"]["endpoint"] == "v1/logs"
    assert event["meta"]["hosts"] == "web-01"


def test_parse_metrics_proto():
    """Test parsing a histogram metric from OTLP protobuf."""
    # Create protobuf message from JSON
    proto_msg = Parse(json.dumps(SAMPLE_METRIC_JSON), ExportMetricsServiceRequest())
    proto_bytes = proto_msg.SerializeToString()

    # Parse protobuf bytes
    result = parse_metrics_proto(proto_bytes)

    assert len(result) == 1
    event = result[0]

    # Check signal type
    assert event["signal_type"] == "metric"

    # Check resource attributes
    assert event["resource"]["host.name"] == "web-01"

    # Check metric data
    metric = event["metric"]
    assert metric["name"] == "http.server.request.duration"
    assert metric["type"] == "histogram"

    # Check metadata
    assert event["meta"]["endpoint"] == "v1/metrics"
    assert event["meta"]["hosts"] == "web-01"


def test_parse_traces_proto():
    """Test parsing a span from OTLP protobuf."""
    # Create protobuf message from JSON
    proto_msg = Parse(json.dumps(SAMPLE_TRACE_JSON), ExportTraceServiceRequest())
    proto_bytes = proto_msg.SerializeToString()

    # Parse protobuf bytes
    result = parse_traces_proto(proto_bytes)

    assert len(result) == 1
    event = result[0]

    # Check signal type
    assert event["signal_type"] == "trace"

    # Check resource attributes
    assert event["resource"]["host.name"] == "web-01"

    # Check span data
    span = event["span"]
    assert span["trace_id"] == "abcdef1234567890abcdef1234567890"
    assert span["name"] == "HTTP GET /api/users"

    # Check metadata
    assert event["meta"]["endpoint"] == "v1/traces"
    assert event["meta"]["hosts"] == "web-01"
