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
