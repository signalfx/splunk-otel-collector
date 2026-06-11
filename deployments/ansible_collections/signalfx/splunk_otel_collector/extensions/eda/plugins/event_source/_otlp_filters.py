"""Event filters for OTLP plugin."""

from __future__ import annotations

import re

# OTel severity number ranges per spec
SEVERITY_NAME_TO_MIN_NUMBER = {
    "TRACE": 1,
    "DEBUG": 5,
    "INFO": 9,
    "WARN": 13,
    "ERROR": 17,
    "FATAL": 21,
}


class SeverityFilter:
    """Filter log events by minimum severity level."""

    def __init__(self, severity_min: str | None) -> None:
        """Initialize severity filter.

        Args:
            severity_min: Minimum severity name (TRACE/DEBUG/INFO/WARN/ERROR/FATAL)
                         or None to accept all events.

        Raises:
            ValueError: If severity_min is unknown.
        """
        if severity_min is None:
            self.threshold = 0
        else:
            if severity_min not in SEVERITY_NAME_TO_MIN_NUMBER:
                raise ValueError(
                    f"Unknown severity: {severity_min}. "
                    f"Valid values: {', '.join(SEVERITY_NAME_TO_MIN_NUMBER.keys())}"
                )
            self.threshold = SEVERITY_NAME_TO_MIN_NUMBER[severity_min]

    def accepts(self, event: dict) -> bool:
        """Check if event passes severity filter.

        Args:
            event: Event dict with signal_type and log/metric/trace data.

        Returns:
            True if event passes filter, False otherwise.
        """
        # Non-log events always pass
        if event.get("signal_type") != "log":
            return True

        # Check log severity
        severity_number = event.get("log", {}).get("severity_number", 0)
        return severity_number >= self.threshold


class ResourceFilter:
    """Filter events by resource attribute patterns."""

    def __init__(self, patterns: dict[str, str] | None) -> None:
        """Initialize resource filter.

        Args:
            patterns: Dict mapping resource attribute names to regex patterns,
                     or None to accept all events.
        """
        self.compiled_patterns: list[tuple[str, re.Pattern]] = []

        if patterns:
            for key, pattern in patterns.items():
                # Anchor pattern to match entire value
                anchored = f"^(?:{pattern})$"
                self.compiled_patterns.append((key, re.compile(anchored)))

    def accepts(self, event: dict) -> bool:
        """Check if event passes resource filter.

        Args:
            event: Event dict with resource attributes.

        Returns:
            True if event passes filter (all patterns match), False otherwise.
        """
        # No patterns = accept all
        if not self.compiled_patterns:
            return True

        resource = event.get("resource", {})

        # All patterns must match
        for key, pattern in self.compiled_patterns:
            value = resource.get(key)
            if value is None or not pattern.match(value):
                return False

        return True
