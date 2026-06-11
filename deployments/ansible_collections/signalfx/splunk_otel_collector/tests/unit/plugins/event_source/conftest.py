"""Shared fixtures for EDA event source plugin tests."""

import asyncio
import json

import pytest


class MockQueue:
    """Mock asyncio.Queue that collects events for assertion."""

    def __init__(self):
        self.events = []

    async def put(self, event):
        self.events = self.events + [event]


@pytest.fixture
def mock_queue():
    return MockQueue()


SAMPLE_LOG_JSON = {
    "resourceLogs": [
        {
            "resource": {
                "attributes": [
                    {"key": "host.id", "value": {"stringValue": "machine-id-001"}},
                    {"key": "host.name", "value": {"stringValue": "web-01"}},
                    {"key": "service.name", "value": {"stringValue": "nginx"}},
                    {"key": "service.version", "value": {"stringValue": "1.25"}},
                    {
                        "key": "deployment.environment.name",
                        "value": {"stringValue": "production"},
                    },
                ]
            },
            "scopeLogs": [
                {
                    "logRecords": [
                        {
                            "timeUnixNano": "1718100600000000000",
                            "severityNumber": 17,
                            "severityText": "ERROR",
                            "body": {
                                "stringValue": "Connection refused to upstream"
                            },
                            "attributes": [
                                {
                                    "key": "error.type",
                                    "value": {
                                        "stringValue": "ConnectionError"
                                    },
                                }
                            ],
                        }
                    ]
                }
            ],
        }
    ]
}

SAMPLE_METRIC_JSON = {
    "resourceMetrics": [
        {
            "resource": {
                "attributes": [
                    {"key": "host.id", "value": {"stringValue": "machine-id-001"}},
                    {"key": "host.name", "value": {"stringValue": "web-01"}},
                    {"key": "service.name", "value": {"stringValue": "nginx"}},
                ]
            },
            "scopeMetrics": [
                {
                    "metrics": [
                        {
                            "name": "http.server.request.duration",
                            "unit": "s",
                            "histogram": {
                                "dataPoints": [
                                    {
                                        "attributes": [
                                            {
                                                "key": "http.request.method",
                                                "value": {
                                                    "stringValue": "GET"
                                                },
                                            }
                                        ],
                                        "startTimeUnixNano": "1718100540000000000",
                                        "timeUnixNano": "1718100600000000000",
                                        "count": "150",
                                        "sum": 45.2,
                                        "bucketCounts": [
                                            "10",
                                            "50",
                                            "60",
                                            "25",
                                            "5",
                                        ],
                                        "explicitBounds": [
                                            0.01,
                                            0.05,
                                            0.1,
                                            0.5,
                                        ],
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

SAMPLE_TRACE_JSON = {
    "resourceSpans": [
        {
            "resource": {
                "attributes": [
                    {"key": "host.id", "value": {"stringValue": "machine-id-001"}},
                    {"key": "host.name", "value": {"stringValue": "web-01"}},
                    {"key": "service.name", "value": {"stringValue": "api-gateway"}},
                ]
            },
            "scopeSpans": [
                {
                    "spans": [
                        {
                            "traceId": "abcdef1234567890abcdef1234567890",
                            "spanId": "1234567890abcdef",
                            "parentSpanId": "",
                            "name": "HTTP GET /api/users",
                            "kind": 2,
                            "startTimeUnixNano": "1718100600000000000",
                            "endTimeUnixNano": "1718100601000000000",
                            "status": {
                                "code": 2,
                                "message": "Internal Server Error",
                            },
                            "attributes": [
                                {
                                    "key": "http.response.status_code",
                                    "value": {"intValue": "500"},
                                },
                                {
                                    "key": "http.request.method",
                                    "value": {"stringValue": "GET"},
                                },
                            ],
                        }
                    ]
                }
            ],
        }
    ]
}
