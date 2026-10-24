"""Strict end-to-end checks for the deterministic file-input test TA."""

from tests.lib import assertions


SOURCETYPE = "otel_uf_test"


def test_exact_count_parity(splunk, uf_index, otel_index):
    """The same appended test lines must produce the same non-zero count."""
    assertions.exact_count_parity(
        splunk, uf_index, otel_index, SOURCETYPE, earliest="-24h", latest="now"
    )


def test_field_parity(splunk, uf_index, otel_index):
    """The deterministic TA must preserve extracted key/value fields."""
    assertions.fields_parity(
        splunk, uf_index, otel_index, SOURCETYPE, earliest="-24h", latest="now"
    )


def test_timestamp_delta(splunk, uf_index, otel_index):
    """The deterministic TA must not introduce a large indexing delay."""
    assertions.timestamp_delta(
        splunk, uf_index, otel_index, SOURCETYPE, earliest="-24h", latest="now"
    )

