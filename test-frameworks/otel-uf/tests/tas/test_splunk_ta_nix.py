"""
Correctness tests for Splunk_TA_nix: OTel Collector vs Universal Forwarder.

Compares data in the UF index against the OTel index across:
  - sourcetype presence
  - event counts (within 5% tolerance)
  - field presence
  - timestamp accuracy
"""
import pytest
from tests.lib import assertions

# Sourcetypes enabled in tas/Splunk_TA_nix/local/inputs.conf
SOURCETYPES = [
    "syslog",
    "linux_secure",
    "bash_history",
]


# ---------------------------------------------------------------------------
# Sourcetype parity (one test covers all sourcetypes at once)
# ---------------------------------------------------------------------------

def test_sourcetype_parity(splunk, uf_index, otel_index):
    """Every sourcetype in the UF index must also appear in the OTel index."""
    assertions.sourcetype_parity(splunk, uf_index, otel_index)


# ---------------------------------------------------------------------------
# Per-sourcetype tests
# ---------------------------------------------------------------------------

@pytest.mark.parametrize("sourcetype", SOURCETYPES)
def test_count_parity(splunk, uf_index, otel_index, sourcetype):
    """Event counts must be within 5% between UF and OTel indexes."""
    assertions.count_parity(splunk, uf_index, otel_index, sourcetype, tolerance=0.05)


@pytest.mark.parametrize("sourcetype", SOURCETYPES)
def test_field_parity(splunk, uf_index, otel_index, sourcetype):
    """Fields present in >50% of UF events must also appear in OTel events."""
    assertions.fields_parity(splunk, uf_index, otel_index, sourcetype)


@pytest.mark.parametrize("sourcetype", SOURCETYPES)
def test_timestamp_delta(splunk, uf_index, otel_index, sourcetype):
    """avg(_indextime - _time) must not exceed 60s in either index."""
    assertions.timestamp_delta(splunk, uf_index, otel_index, sourcetype)
