#!/usr/bin/env python3
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)
"""
OTLP/HTTP event source plugin for Ansible EDA.

Receives OpenTelemetry Protocol (OTLP) signals via HTTP and emits EDA events.
Supports logs, metrics, and traces in JSON and protobuf formats.
"""

from __future__ import annotations

DOCUMENTATION = r"""
---
name: otlp
short_description: Receive OpenTelemetry signals via OTLP/HTTP
description:
  - Runs an HTTP server that accepts OTLP/HTTP requests.
  - Supports logs, metrics, and traces in JSON and protobuf formats.
  - Can filter events by signal type, severity, and resource attributes.
  - Supports TLS and bearer token authentication.
version_added: "1.0.0"
author:
  - SignalFx EDA Team
options:
  host:
    description:
      - IP address to bind the HTTP server to.
    type: str
    default: "0.0.0.0"
  port:
    description:
      - TCP port to bind the HTTP server to.
    type: int
    required: true
  signal_types:
    description:
      - List of signal types to accept (logs, metrics, traces).
      - If not specified, all signal types are accepted.
    type: list
    elements: str
    choices: ['logs', 'metrics', 'traces']
  severity_min:
    description:
      - Minimum severity level for log events (TRACE, DEBUG, INFO, WARN, ERROR, FATAL).
      - Only applies to log events. Metrics and traces always pass.
    type: str
    choices: ['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL']
  resource_filters:
    description:
      - Dictionary mapping resource attribute names to regex patterns.
      - All patterns must match for an event to be accepted.
    type: dict
  certfile:
    description:
      - Path to TLS certificate file for HTTPS.
      - Requires keyfile to be set.
    type: str
  keyfile:
    description:
      - Path to TLS private key file for HTTPS.
      - Requires certfile to be set.
    type: str
  cafile:
    description:
      - Path to CA certificate file for client verification.
      - Optional. Only used with certfile and keyfile.
    type: str
  token:
    description:
      - Bearer token for authentication.
      - If set, all requests must include "Authorization: Bearer <token>" header.
    type: str
"""

EXAMPLES = r"""
---
- name: Accept all OTLP signals on port 4318
  signalfx.splunk_otel_collector.otlp:
    host: 0.0.0.0
    port: 4318

- name: Accept only ERROR and FATAL logs
  signalfx.splunk_otel_collector.otlp:
    host: 127.0.0.1
    port: 4318
    signal_types:
      - logs
    severity_min: ERROR

- name: Filter by resource attributes
  signalfx.splunk_otel_collector.otlp:
    host: 0.0.0.0
    port: 4318
    resource_filters:
      deployment.environment.name: "production|staging"
      service.name: "api-.*"

- name: Enable TLS and bearer auth
  signalfx.splunk_otel_collector.otlp:
    host: 0.0.0.0
    port: 4318
    certfile: /etc/ssl/certs/server.crt
    keyfile: /etc/ssl/private/server.key
    cafile: /etc/ssl/certs/ca.crt
    token: "{{ lookup('env', 'OTLP_TOKEN') }}"
"""

import asyncio
import gzip
import json
import ssl
from typing import Any

from aiohttp import web

from ._otlp_filters import ResourceFilter, SeverityFilter
from ._otlp_parsers import (
    parse_logs_json,
    parse_logs_proto,
    parse_metrics_json,
    parse_metrics_proto,
    parse_traces_json,
    parse_traces_proto,
)

# Signal endpoint mapping
SIGNAL_ENDPOINT_MAP = {
    "v1/logs": "logs",
    "v1/metrics": "metrics",
    "v1/traces": "traces",
}

# Parser mapping
JSON_PARSERS = {
    "logs": parse_logs_json,
    "metrics": parse_metrics_json,
    "traces": parse_traces_json,
}

PROTO_PARSERS = {
    "logs": parse_logs_proto,
    "metrics": parse_metrics_proto,
    "traces": parse_traces_proto,
}


def bearer_auth(token: str):
    """Create bearer token authentication middleware.

    Args:
        token: Expected bearer token value.

    Returns:
        aiohttp middleware function.
    """

    @web.middleware
    async def middleware(request: web.Request, handler):
        """Check Authorization header for bearer token."""
        auth_header = request.headers.get("Authorization", "")

        if not auth_header.startswith("Bearer "):
            return web.Response(status=401, text="Unauthorized")

        request_token = auth_header[7:]  # Strip "Bearer "
        if request_token != token:
            return web.Response(status=401, text="Unauthorized")

        return await handler(request)

    return middleware


def _build_ssl_context(args: dict) -> ssl.SSLContext | None:
    """Build SSL context from plugin arguments.

    Args:
        args: Plugin arguments dict.

    Returns:
        SSLContext if certfile provided, None otherwise.
    """
    certfile = args.get("certfile")
    if not certfile:
        return None

    keyfile = args.get("keyfile")
    cafile = args.get("cafile")

    context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
    context.load_cert_chain(certfile=certfile, keyfile=keyfile)

    if cafile:
        context.load_verify_locations(cafile=cafile)
        context.verify_mode = ssl.CERT_REQUIRED

    return context


async def _handle_otlp(request: web.Request) -> web.Response:
    """Handle OTLP/HTTP POST requests.

    Args:
        request: aiohttp request object.

    Returns:
        HTTP response (200 on success, 400/404 on error).
    """
    # Extract signal type from URL path
    signal_path = request.match_info.get("signal_path", "")
    endpoint = f"v1/{signal_path}"
    signal_type = SIGNAL_ENDPOINT_MAP.get(endpoint)

    if not signal_type:
        return web.Response(status=404, text="Not Found")

    # Check if signal type is allowed
    allowed_signals = request.app["allowed_signals"]
    if signal_type not in allowed_signals:
        return web.Response(status=404, text="Not Found")

    # Read request body
    try:
        body = await request.read()
    except Exception:
        return web.Response(status=400, text="Bad Request")

    # Decompress if gzip (check both header and magic bytes)
    content_encoding = request.headers.get("Content-Encoding", "")
    is_gzipped = content_encoding == "gzip" or (len(body) >= 2 and body[:2] == b'\x1f\x8b')

    if is_gzipped:
        try:
            body = gzip.decompress(body)
        except Exception:
            return web.Response(status=400, text="Invalid gzip")

    # Parse based on Content-Type
    content_type = request.headers.get("Content-Type", "")

    try:
        if "application/x-protobuf" in content_type:
            parser = PROTO_PARSERS[signal_type]
            events = parser(body)
        else:
            # Default to JSON
            payload = json.loads(body)
            parser = JSON_PARSERS[signal_type]
            events = parser(payload)
    except (json.JSONDecodeError, Exception):
        return web.Response(status=400, text="Invalid payload")

    # Apply filters and enqueue events
    severity_filter = request.app["severity_filter"]
    resource_filter = request.app["resource_filter"]
    queue = request.app["queue"]

    for event in events:
        if severity_filter.accepts(event) and resource_filter.accepts(event):
            await queue.put(event)

    return web.json_response({})


async def main(queue: Any, args: dict) -> None:
    """Main entry point for OTLP event source plugin.

    Args:
        queue: asyncio.Queue for emitting events.
        args: Plugin arguments dict.
    """
    # Extract configuration
    host = args.get("host", "0.0.0.0")
    port = args["port"]  # Required
    signal_types = args.get("signal_types")
    severity_min = args.get("severity_min")
    resource_filters = args.get("resource_filters")
    token = args.get("token")

    # Build allowed signals set
    if signal_types:
        allowed_signals = set(signal_types)
    else:
        allowed_signals = {"logs", "metrics", "traces"}

    # Build filters
    severity_filter = SeverityFilter(severity_min)
    resource_filter = ResourceFilter(resource_filters)

    # Create aiohttp app
    app = web.Application()

    # Add bearer auth middleware if token provided
    if token:
        app.middlewares.append(bearer_auth(token))

    # Store app state
    app["queue"] = queue
    app["allowed_signals"] = allowed_signals
    app["severity_filter"] = severity_filter
    app["resource_filter"] = resource_filter

    # Register route
    app.router.add_post("/v1/{signal_path:logs|metrics|traces}", _handle_otlp)

    # Build SSL context
    ssl_context = _build_ssl_context(args)

    # Start server
    runner = web.AppRunner(app)
    await runner.setup()

    site = web.TCPSite(runner, host, port, ssl_context=ssl_context)
    await site.start()

    # Run until cancelled
    try:
        await asyncio.Future()
    except asyncio.CancelledError:
        pass
    finally:
        await runner.cleanup()


class MockQueue:
    """Mock queue for standalone testing."""

    async def put(self, event: dict) -> None:
        """Print event to stdout."""
        print(json.dumps(event, indent=2))


if __name__ == "__main__":
    # Standalone testing mode
    test_args = {
        "host": "127.0.0.1",
        "port": 4318,
    }

    test_queue = MockQueue()

    print(f"Starting OTLP server on {test_args['host']}:{test_args['port']}")
    print("Press Ctrl+C to stop")

    try:
        asyncio.run(main(test_queue, test_args))
    except KeyboardInterrupt:
        print("\nShutting down...")
