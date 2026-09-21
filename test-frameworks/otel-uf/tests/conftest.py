"""
Shared pytest fixtures for OTel vs UF correctness tests.

Required environment variables (or passed via pytest CLI options):
  SPLUNK_HOST, SPLUNK_PORT, SPLUNK_PASSWORD, UF_INDEX, OTEL_INDEX

CLI options (override env vars):
  --splunk-host, --port, --password, --uf-index, --otel-index
"""
import os
import pytest
import splunklib.client as splunk_client


def pytest_addoption(parser):
    parser.addoption("--splunk-host", default=os.environ.get("SPLUNK_HOST", "localhost"))
    parser.addoption("--port", type=int, default=int(os.environ.get("SPLUNK_MGMT_PORT", "8089")))
    parser.addoption("--username", default=os.environ.get("SPLUNK_ADMIN_USER", "admin"))
    parser.addoption("--password", default=os.environ.get("SPLUNK_ADMIN_PASSWORD", ""))
    parser.addoption("--uf-index", default=os.environ.get("UF_INDEX", "uf_nix"))
    parser.addoption("--otel-index", default=os.environ.get("OTEL_INDEX", "otel_nix"))


@pytest.fixture(scope="session")
def splunk(request):
    """Authenticated splunk-sdk Service instance (session-scoped)."""
    service = splunk_client.connect(
        host=request.config.getoption("--splunk-host"),
        port=request.config.getoption("--port"),
        username=request.config.getoption("--username"),
        password=request.config.getoption("--password"),
    )
    return service


@pytest.fixture(scope="session")
def uf_index(request):
    return request.config.getoption("--uf-index")


@pytest.fixture(scope="session")
def otel_index(request):
    return request.config.getoption("--otel-index")
