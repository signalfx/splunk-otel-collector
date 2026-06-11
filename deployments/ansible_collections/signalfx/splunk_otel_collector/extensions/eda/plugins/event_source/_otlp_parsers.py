"""OTLP payload parsers for converting JSON to EDA events."""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any


def extract_resource_attributes(attrs: list[dict]) -> dict[str, Any]:
    """
    Convert OTLP attribute list to flat dictionary.

    Args:
        attrs: List of OTLP attributes with {"key": "...", "value": {...}}

    Returns:
        Flat dictionary with extracted values
    """
    result = {}
    for attr in attrs:
        key = attr.get("key")
        value_obj = attr.get("value", {})
        if key:
            result[key] = _extract_any_value(value_obj)
    return result


def _extract_any_value(value_obj: dict) -> Any:
    """
    Extract value from OTLP AnyValue union type.

    Handles: stringValue, intValue, doubleValue, boolValue,
    arrayValue, kvlistValue, bytesValue.

    Args:
        value_obj: OTLP value object

    Returns:
        Extracted Python value
    """
    if "stringValue" in value_obj:
        return value_obj["stringValue"]
    elif "intValue" in value_obj:
        # intValue is a string in JSON, convert to int
        return int(value_obj["intValue"])
    elif "doubleValue" in value_obj:
        return value_obj["doubleValue"]
    elif "boolValue" in value_obj:
        return value_obj["boolValue"]
    elif "arrayValue" in value_obj:
        array_values = value_obj["arrayValue"].get("values", [])
        return [_extract_any_value(v) for v in array_values]
    elif "kvlistValue" in value_obj:
        kv_values = value_obj["kvlistValue"].get("values", [])
        return extract_resource_attributes(kv_values)
    elif "bytesValue" in value_obj:
        return value_obj["bytesValue"]
    else:
        return None


def nanos_to_iso(nanos_str: str) -> str:
    """
    Convert nanosecond unix timestamp string to ISO 8601 UTC.

    Args:
        nanos_str: Nanosecond unix timestamp as string

    Returns:
        ISO 8601 formatted timestamp or empty string for "0"
    """
    if nanos_str == "0":
        return ""

    nanos = int(nanos_str)
    seconds = nanos / 1_000_000_000
    dt = datetime.fromtimestamp(seconds, tz=timezone.utc)
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")


def _extract_body(body_obj: dict | None) -> Any:
    """
    Extract log body from OTLP AnyValue.

    Args:
        body_obj: OTLP body object or None

    Returns:
        Extracted body value or empty string
    """
    if not body_obj:
        return ""
    return _extract_any_value(body_obj)


def _parse_metric_data_points(metric: dict) -> tuple[str, list[dict]]:
    """
    Extract metric type and data points from OTLP metric.

    Args:
        metric: OTLP metric object

    Returns:
        Tuple of (type_name, data_points_list)
    """
    # Determine metric type
    metric_types = [
        "gauge",
        "sum",
        "histogram",
        "exponentialHistogram",
        "summary",
    ]

    type_name = ""
    metric_data = {}

    for mtype in metric_types:
        if mtype in metric:
            # Normalize exponentialHistogram to snake_case
            type_name = (
                "exponential_histogram"
                if mtype == "exponentialHistogram"
                else mtype
            )
            metric_data = metric[mtype]
            break

    if not type_name:
        return ("", [])

    # Extract data points
    data_points = []
    for dp in metric_data.get("dataPoints", []):
        point = {
            "attributes": extract_resource_attributes(dp.get("attributes", [])),
            "start_time": nanos_to_iso(dp.get("startTimeUnixNano", "0")),
            "time": nanos_to_iso(dp.get("timeUnixNano", "0")),
        }

        # Extract value (asDouble or asInt)
        if "asDouble" in dp:
            point["value"] = dp["asDouble"]
        elif "asInt" in dp:
            point["value"] = int(dp["asInt"])

        # For histograms, add bucket data
        if type_name in ("histogram", "exponential_histogram"):
            if "count" in dp:
                point["count"] = int(dp["count"])
            if "sum" in dp:
                point["sum"] = dp["sum"]
            if "bucketCounts" in dp:
                point["bucket_counts"] = [int(c) for c in dp["bucketCounts"]]
            if "explicitBounds" in dp:
                point["explicit_bounds"] = dp["explicitBounds"]

        data_points.append(point)

    return (type_name, data_points)


def parse_metrics_json(payload: dict) -> list[dict]:
    """
    Parse OTLP ExportMetricsServiceRequest JSON into EDA events.

    Each metric becomes one event with structure:
    {
        "signal_type": "metric",
        "resource": {extracted resource attributes},
        "metric": {
            "name": str,
            "type": str,
            "unit": str,
            "data_points": [...],
        },
        "meta": {
            "endpoint": "v1/metrics",
            "hosts": str (from resource["host.name"], fallback "localhost"),
        },
    }

    Args:
        payload: OTLP ExportMetricsServiceRequest as dict

    Returns:
        List of EDA event dictionaries
    """
    events = []

    for resource_metric in payload.get("resourceMetrics", []):
        resource_attrs = extract_resource_attributes(
            resource_metric.get("resource", {}).get("attributes", [])
        )

        # Extract host name for meta.hosts
        host_name = resource_attrs.get("host.name", "localhost")

        for scope_metric in resource_metric.get("scopeMetrics", []):
            for metric in scope_metric.get("metrics", []):
                # Parse metric data points
                metric_type, data_points = _parse_metric_data_points(metric)

                # Build event
                event = {
                    "signal_type": "metric",
                    "resource": resource_attrs,
                    "metric": {
                        "name": metric.get("name", ""),
                        "type": metric_type,
                        "unit": metric.get("unit", ""),
                        "data_points": data_points,
                    },
                    "meta": {
                        "endpoint": "v1/metrics",
                        "hosts": host_name,
                    },
                }
                events.append(event)

    return events


def parse_traces_json(payload: dict) -> list[dict]:
    """
    Parse OTLP ExportTraceServiceRequest JSON into EDA events.

    Each span becomes one event with structure:
    {
        "signal_type": "trace",
        "resource": {extracted resource attributes},
        "span": {
            "trace_id": str,
            "span_id": str,
            "parent_span_id": str,
            "name": str,
            "kind": int,
            "start_time": str,
            "end_time": str,
            "status_code": int,
            "status_message": str,
            "attributes": dict,
        },
        "meta": {
            "endpoint": "v1/traces",
            "hosts": str (from resource["host.name"], fallback "localhost"),
        },
    }

    Args:
        payload: OTLP ExportTraceServiceRequest as dict

    Returns:
        List of EDA event dictionaries
    """
    events = []

    for resource_span in payload.get("resourceSpans", []):
        resource_attrs = extract_resource_attributes(
            resource_span.get("resource", {}).get("attributes", [])
        )

        # Extract host name for meta.hosts
        host_name = resource_attrs.get("host.name", "localhost")

        for scope_span in resource_span.get("scopeSpans", []):
            for span in scope_span.get("spans", []):
                # Extract span attributes
                span_attrs = extract_resource_attributes(span.get("attributes", []))

                # Extract status
                status = span.get("status", {})

                # Build event
                event = {
                    "signal_type": "trace",
                    "resource": resource_attrs,
                    "span": {
                        "trace_id": span.get("traceId", ""),
                        "span_id": span.get("spanId", ""),
                        "parent_span_id": span.get("parentSpanId", ""),
                        "name": span.get("name", ""),
                        "kind": span.get("kind", 0),
                        "start_time": nanos_to_iso(
                            span.get("startTimeUnixNano", "0")
                        ),
                        "end_time": nanos_to_iso(span.get("endTimeUnixNano", "0")),
                        "status_code": status.get("code", 0),
                        "status_message": status.get("message", ""),
                        "attributes": span_attrs,
                    },
                    "meta": {
                        "endpoint": "v1/traces",
                        "hosts": host_name,
                    },
                }
                events.append(event)

    return events


def parse_logs_json(payload: dict) -> list[dict]:
    """
    Parse OTLP ExportLogsServiceRequest JSON into EDA events.

    Each log record becomes one event with structure:
    {
        "signal_type": "log",
        "resource": {extracted resource attributes},
        "log": {
            "severity_text": str,
            "severity_number": int,
            "body": str,
            "attributes": dict,
            "timestamp": str (ISO 8601),
        },
        "meta": {
            "endpoint": "v1/logs",
            "hosts": str (from resource["host.name"], fallback "localhost"),
        },
    }

    Args:
        payload: OTLP ExportLogsServiceRequest as dict

    Returns:
        List of EDA event dictionaries
    """
    events = []

    for resource_log in payload.get("resourceLogs", []):
        resource_attrs = extract_resource_attributes(
            resource_log.get("resource", {}).get("attributes", [])
        )

        # Extract host name for meta.hosts
        host_name = resource_attrs.get("host.name", "localhost")

        for scope_log in resource_log.get("scopeLogs", []):
            for log_record in scope_log.get("logRecords", []):
                # Extract log attributes
                log_attrs = extract_resource_attributes(
                    log_record.get("attributes", [])
                )

                # Build event
                event = {
                    "signal_type": "log",
                    "resource": resource_attrs,
                    "log": {
                        "severity_text": log_record.get("severityText", ""),
                        "severity_number": log_record.get("severityNumber", 0),
                        "body": _extract_body(log_record.get("body")),
                        "attributes": log_attrs,
                        "timestamp": nanos_to_iso(
                            log_record.get("timeUnixNano", "0")
                        ),
                    },
                    "meta": {
                        "endpoint": "v1/logs",
                        "hosts": host_name,
                    },
                }
                events.append(event)

    return events
