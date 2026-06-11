"""Tests for OTLP event source plugin HTTP server."""

import asyncio
import gzip
import json

import aiohttp
import pytest
import pytest_asyncio
from google.protobuf.json_format import Parse
from opentelemetry.proto.collector.logs.v1.logs_service_pb2 import (
    ExportLogsServiceRequest,
)

from .conftest import SAMPLE_LOG_JSON, SAMPLE_METRIC_JSON, SAMPLE_TRACE_JSON, MockQueue


@pytest_asyncio.fixture
async def running_plugin():
    """Start OTLP plugin server for testing."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14318}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)
    yield queue, 14318, task
    task.cancel()
    try:
        await task
    except asyncio.CancelledError:
        pass


@pytest.mark.asyncio
async def test_server_starts_and_accepts_logs(running_plugin):
    """Test server starts and accepts log payload."""
    queue, port, _ = running_plugin

    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"http://127.0.0.1:{port}/v1/logs",
            json=SAMPLE_LOG_JSON,
            headers={"Content-Type": "application/json"},
        ) as resp:
            assert resp.status == 200

    # Should have one event on queue
    assert len(queue.events) == 1
    assert queue.events[0]["signal_type"] == "log"
    assert queue.events[0]["log"]["severity_text"] == "ERROR"


@pytest.mark.asyncio
async def test_server_accepts_metrics(running_plugin):
    """Test server accepts metric payload."""
    queue, port, _ = running_plugin

    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"http://127.0.0.1:{port}/v1/metrics",
            json=SAMPLE_METRIC_JSON,
            headers={"Content-Type": "application/json"},
        ) as resp:
            assert resp.status == 200

    # Should have one event on queue
    assert len(queue.events) == 1
    assert queue.events[0]["signal_type"] == "metric"
    assert queue.events[0]["metric"]["name"] == "http.server.request.duration"


@pytest.mark.asyncio
async def test_server_accepts_traces(running_plugin):
    """Test server accepts trace payload."""
    queue, port, _ = running_plugin

    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"http://127.0.0.1:{port}/v1/traces",
            json=SAMPLE_TRACE_JSON,
            headers={"Content-Type": "application/json"},
        ) as resp:
            assert resp.status == 200

    # Should have one event on queue
    assert len(queue.events) == 1
    assert queue.events[0]["signal_type"] == "trace"
    assert queue.events[0]["span"]["name"] == "HTTP GET /api/users"


@pytest.mark.asyncio
async def test_server_rejects_get():
    """Test server rejects GET requests."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14319}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"http://127.0.0.1:14319/v1/logs") as resp:
                assert resp.status == 405
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


@pytest.mark.asyncio
async def test_server_rejects_invalid_json():
    """Test server rejects invalid JSON payload."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14320}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"http://127.0.0.1:14320/v1/logs",
                data="not valid json",
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 400
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


@pytest.mark.asyncio
async def test_server_rejects_unknown_endpoint():
    """Test server rejects unknown endpoint."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14321}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"http://127.0.0.1:14321/v1/unknown",
                json={"test": "data"},
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 404
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


# Task 7: Signal Type Filtering and Bearer Auth Tests


@pytest.mark.asyncio
async def test_signal_types_filter_rejects_disabled():
    """Test signal_types filter rejects disabled signal types."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14322, "signal_types": ["logs"]}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            # Traces endpoint should be rejected
            async with session.post(
                f"http://127.0.0.1:14322/v1/traces",
                json=SAMPLE_TRACE_JSON,
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 404

            # Logs endpoint should work
            async with session.post(
                f"http://127.0.0.1:14322/v1/logs",
                json=SAMPLE_LOG_JSON,
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 200
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


@pytest.mark.asyncio
async def test_bearer_auth_accepts_valid_token():
    """Test bearer auth accepts valid token."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14323, "token": "secret123"}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"http://127.0.0.1:14323/v1/logs",
                json=SAMPLE_LOG_JSON,
                headers={
                    "Content-Type": "application/json",
                    "Authorization": "Bearer secret123",
                },
            ) as resp:
                assert resp.status == 200
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


@pytest.mark.asyncio
async def test_bearer_auth_rejects_missing_token():
    """Test bearer auth rejects requests without token."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14324, "token": "secret123"}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"http://127.0.0.1:14324/v1/logs",
                json=SAMPLE_LOG_JSON,
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 401
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


@pytest.mark.asyncio
async def test_bearer_auth_rejects_wrong_token():
    """Test bearer auth rejects requests with wrong token."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14325, "token": "secret123"}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"http://127.0.0.1:14325/v1/logs",
                json=SAMPLE_LOG_JSON,
                headers={
                    "Content-Type": "application/json",
                    "Authorization": "Bearer wrongtoken",
                },
            ) as resp:
                assert resp.status == 401
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


# Task 8: Severity and Resource Filter Integration Tests


@pytest.mark.asyncio
async def test_severity_filter_drops_below_threshold():
    """Test severity_min filter drops logs below threshold."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {"host": "127.0.0.1", "port": 14326, "severity_min": "ERROR"}
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            # Send INFO log (severityNumber=9) - should be dropped
            info_log = {
                "resourceLogs": [
                    {
                        "resource": {"attributes": []},
                        "scopeLogs": [
                            {
                                "logRecords": [
                                    {
                                        "timeUnixNano": "1718100600000000000",
                                        "severityNumber": 9,
                                        "severityText": "INFO",
                                        "body": {"stringValue": "Info message"},
                                        "attributes": [],
                                    }
                                ]
                            }
                        ],
                    }
                ]
            }
            async with session.post(
                f"http://127.0.0.1:14326/v1/logs",
                json=info_log,
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 200
                # Queue should be empty (filtered out)
                assert len(queue.events) == 0

            # Send ERROR log (severityNumber=17) - should be accepted
            async with session.post(
                f"http://127.0.0.1:14326/v1/logs",
                json=SAMPLE_LOG_JSON,
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 200
                # Queue should have 1 event
                assert len(queue.events) == 1
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


@pytest.mark.asyncio
async def test_resource_filter_drops_non_matching():
    """Test resource_filters drops non-matching resources."""
    from extensions.eda.plugins.event_source.otlp import main

    queue = MockQueue()
    args = {
        "host": "127.0.0.1",
        "port": 14327,
        "resource_filters": {"service.name": "redis"},
    }
    task = asyncio.create_task(main(queue, args))
    await asyncio.sleep(0.3)

    try:
        async with aiohttp.ClientSession() as session:
            # Send nginx log (SAMPLE_LOG_JSON has service.name=nginx)
            # Should be rejected since filter expects redis
            async with session.post(
                f"http://127.0.0.1:14327/v1/logs",
                json=SAMPLE_LOG_JSON,
                headers={"Content-Type": "application/json"},
            ) as resp:
                assert resp.status == 200
                # Queue should be empty (filtered out)
                assert len(queue.events) == 0
    finally:
        task.cancel()
        try:
            await task
        except asyncio.CancelledError:
            pass


# Task 9: Gzip and Protobuf Tests


@pytest.mark.asyncio
async def test_gzip_decompression(running_plugin):
    """Test server decompresses gzip-encoded payloads."""
    queue, port, _ = running_plugin

    # Track initial queue length
    initial_length = len(queue.events)

    # Gzip compress the JSON payload
    json_bytes = json.dumps(SAMPLE_LOG_JSON).encode("utf-8")
    compressed = gzip.compress(json_bytes)

    async with aiohttp.ClientSession() as session:
        # Note: Don't set Content-Encoding header here because aiohttp client
        # will decompress the payload before sending. The server auto-detects
        # gzip by magic bytes (0x1f 0x8b).
        async with session.post(
            f"http://127.0.0.1:{port}/v1/logs",
            data=compressed,
            headers={"Content-Type": "application/json"},
        ) as resp:
            assert resp.status == 200

    # Should have one new event on queue
    assert len(queue.events) == initial_length + 1
    latest_event = queue.events[-1]
    assert latest_event["signal_type"] == "log"
    assert latest_event["log"]["severity_text"] == "ERROR"


@pytest.mark.asyncio
async def test_protobuf_content_type(running_plugin):
    """Test server accepts protobuf content type."""
    queue, port, _ = running_plugin

    # Convert SAMPLE_LOG_JSON to protobuf bytes
    proto_request = Parse(json.dumps(SAMPLE_LOG_JSON), ExportLogsServiceRequest())
    proto_bytes = proto_request.SerializeToString()

    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"http://127.0.0.1:{port}/v1/logs",
            data=proto_bytes,
            headers={"Content-Type": "application/x-protobuf"},
        ) as resp:
            assert resp.status == 200

    # Should have one event on queue with correct data
    assert len(queue.events) == 1
    assert queue.events[0]["signal_type"] == "log"
    assert queue.events[0]["log"]["severity_text"] == "ERROR"
    assert queue.events[0]["log"]["body"] == "Connection refused to upstream"
