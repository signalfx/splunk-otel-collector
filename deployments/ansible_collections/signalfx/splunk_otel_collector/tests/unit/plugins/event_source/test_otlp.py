"""Tests for OTLP event source plugin HTTP server."""

import asyncio
import json

import aiohttp
import pytest
import pytest_asyncio

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
