"""Tests for OTLP event filters."""

from __future__ import annotations

import pytest

from extensions.eda.plugins.event_source._otlp_filters import (
    SEVERITY_NAME_TO_MIN_NUMBER,
    ResourceFilter,
    SeverityFilter,
)


def test_severity_name_mapping():
    """Verify all 6 severity level mappings per OTel spec."""
    assert SEVERITY_NAME_TO_MIN_NUMBER == {
        "TRACE": 1,
        "DEBUG": 5,
        "INFO": 9,
        "WARN": 13,
        "ERROR": 17,
        "FATAL": 21,
    }


def test_severity_filter_accepts_above_threshold():
    """WARN filter should accept ERROR severity."""
    severity_filter = SeverityFilter("WARN")
    event = {
        "signal_type": "log",
        "log": {"severity_number": 17},
    }
    assert severity_filter.accepts(event) is True


def test_severity_filter_rejects_below_threshold():
    """WARN filter should reject INFO severity."""
    severity_filter = SeverityFilter("WARN")
    event = {
        "signal_type": "log",
        "log": {"severity_number": 9},
    }
    assert severity_filter.accepts(event) is False


def test_severity_filter_accepts_at_threshold():
    """WARN filter should accept WARN severity."""
    severity_filter = SeverityFilter("WARN")
    event = {
        "signal_type": "log",
        "log": {"severity_number": 13},
    }
    assert severity_filter.accepts(event) is True


def test_severity_filter_skips_non_log_events():
    """ERROR filter should accept metric events."""
    severity_filter = SeverityFilter("ERROR")
    event = {
        "signal_type": "metric",
        "metric": {"name": "cpu.usage"},
    }
    assert severity_filter.accepts(event) is True


def test_severity_filter_none_accepts_all():
    """None filter should accept all events."""
    severity_filter = SeverityFilter(None)
    event = {
        "signal_type": "log",
        "log": {"severity_number": 1},
    }
    assert severity_filter.accepts(event) is True


def test_severity_filter_invalid_name_raises():
    """Invalid severity name should raise ValueError."""
    with pytest.raises(ValueError, match="Unknown severity"):
        SeverityFilter("INVALID")


def test_resource_filter_matches_exact():
    """Filter should match exact service name."""
    resource_filter = ResourceFilter({"service.name": "nginx"})
    event = {
        "resource": {"service.name": "nginx"},
    }
    assert resource_filter.accepts(event) is True


def test_resource_filter_matches_regex():
    """Filter should match regex pattern."""
    resource_filter = ResourceFilter({"service.name": "nginx|redis"})
    event = {
        "resource": {"service.name": "redis"},
    }
    assert resource_filter.accepts(event) is True


def test_resource_filter_rejects_no_match():
    """Filter should reject non-matching value."""
    resource_filter = ResourceFilter({"service.name": "nginx"})
    event = {
        "resource": {"service.name": "postgres"},
    }
    assert resource_filter.accepts(event) is False


def test_resource_filter_rejects_missing_attribute():
    """Filter should reject event missing required attribute."""
    resource_filter = ResourceFilter({"deployment.environment.name": "prod"})
    event = {
        "resource": {"service.name": "nginx"},
    }
    assert resource_filter.accepts(event) is False


def test_resource_filter_multiple_conditions_all_must_match():
    """Filter should require all conditions to match."""
    resource_filter = ResourceFilter({
        "service.name": "nginx",
        "deployment.environment.name": "prod",
    })

    # Both match
    event_match = {
        "resource": {
            "service.name": "nginx",
            "deployment.environment.name": "prod",
        },
    }
    assert resource_filter.accepts(event_match) is True

    # Partial match
    event_partial = {
        "resource": {
            "service.name": "nginx",
            "deployment.environment.name": "dev",
        },
    }
    assert resource_filter.accepts(event_partial) is False


def test_resource_filter_none_accepts_all():
    """None filter should accept all events."""
    resource_filter = ResourceFilter(None)
    event = {
        "resource": {"service.name": "nginx"},
    }
    assert resource_filter.accepts(event) is True
