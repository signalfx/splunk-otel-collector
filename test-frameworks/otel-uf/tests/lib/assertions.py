"""
Reusable assertion helpers that compare a UF index against an OTel index.

All helpers raise AssertionError with a descriptive message on failure so
pytest reports them as test failures with context.
"""
from . import splunk as splunk_lib


# ---------------------------------------------------------------------------
# Sourcetype parity
# ---------------------------------------------------------------------------

def sourcetype_parity(service, uf_index, otel_index, earliest="-24h", latest="now"):
    """
    Assert that every sourcetype present in the UF index also appears in the
    OTel index.  Extra sourcetypes in OTel are allowed.

    Returns (uf_sourcetypes, otel_sourcetypes) for informational use.
    """
    def _sourcetypes(index):
        rows = splunk_lib.search(
            service,
            f"search index={index} | stats count by sourcetype",
            earliest=earliest,
            latest=latest,
        )
        return {r["sourcetype"] for r in rows if r.get("sourcetype")}

    uf_st = _sourcetypes(uf_index)
    otel_st = _sourcetypes(otel_index)

    missing = uf_st - otel_st
    assert not missing, (
        f"Sourcetypes present in UF index '{uf_index}' but missing from OTel index "
        f"'{otel_index}': {sorted(missing)}"
    )
    return uf_st, otel_st


# ---------------------------------------------------------------------------
# Event count parity
# ---------------------------------------------------------------------------

def count_parity(service, uf_index, otel_index, sourcetype, tolerance=0.05,
                 earliest="-24h", latest="now"):
    """
    Assert that event counts for *sourcetype* in both indexes are within
    *tolerance* (default 5%) of each other.

    Skips the check if both counts are zero.
    """
    def _count(index):
        row = splunk_lib.search_one(
            service,
            f"search index={index} sourcetype={sourcetype} | stats count",
            earliest=earliest,
            latest=latest,
        )
        return int(row["count"]) if row else 0

    uf_count = _count(uf_index)
    otel_count = _count(otel_index)

    if uf_count == 0 and otel_count == 0:
        return  # nothing to compare

    assert uf_count > 0, (
        f"UF index '{uf_index}' has 0 events for sourcetype '{sourcetype}'"
    )
    assert otel_count > 0, (
        f"OTel index '{otel_index}' has 0 events for sourcetype '{sourcetype}'"
    )

    ratio = abs(uf_count - otel_count) / max(uf_count, otel_count)
    assert ratio <= tolerance, (
        f"Event count mismatch for sourcetype '{sourcetype}': "
        f"UF={uf_count}, OTel={otel_count}, diff={ratio:.1%} > tolerance={tolerance:.1%}"
    )


def exact_count_parity(service, uf_index, otel_index, sourcetype,
                       earliest="-24h", latest="now"):
    """Require equal, non-zero event counts in the two indexes.

    This is used by the deterministic test TA. Unlike ``count_parity``, a
    zero-versus-zero result is a failure so an un-ingested test event cannot
    silently make the test pass.
    """
    def _count(index):
        row = splunk_lib.search_one(
            service,
            f"search index={index} sourcetype={sourcetype} | stats count",
            earliest=earliest,
            latest=latest,
        )
        return int(row["count"]) if row else 0

    uf_count = _count(uf_index)
    otel_count = _count(otel_index)

    assert uf_count > 0, (
        f"UF index '{uf_index}' has no events for sourcetype '{sourcetype}'. "
        "Append a test event to both agents before running pytest."
    )
    assert otel_count > 0, (
        f"OTel index '{otel_index}' has no events for sourcetype '{sourcetype}'. "
        "Append a test event to both agents before running pytest."
    )
    assert uf_count == otel_count, (
        f"Exact event count mismatch for sourcetype '{sourcetype}': "
        f"UF={uf_count}, OTel={otel_count}"
    )


# ---------------------------------------------------------------------------
# Field presence parity
# ---------------------------------------------------------------------------

def fields_parity(service, uf_index, otel_index, sourcetype, coverage_threshold=0.5,
                  earliest="-24h", latest="now"):
    """
    Assert that every field present in >*coverage_threshold* of UF events for
    *sourcetype* also appears in the OTel index.

    Uses | fieldsummary to get field coverage per index.
    """
    def _fields(index):
        rows = splunk_lib.search(
            service,
            f"search index={index} sourcetype={sourcetype} | fieldsummary",
            earliest=earliest,
            latest=latest,
        )
        # fieldsummary returns field, count, distinct_count, is_exact, min, max, mean, stdev, modes
        # 'coverage' is not a native field - compute from count / total
        total_row = splunk_lib.search_one(
            service,
            f"search index={index} sourcetype={sourcetype} | stats count",
            earliest=earliest,
            latest=latest,
        )
        total = int(total_row["count"]) if total_row else 0
        if total == 0:
            return set()

        significant = set()
        for r in rows:
            field = r.get("field", "")
            count = int(r.get("count", 0))
            if field and not field.startswith("_") and count / total >= coverage_threshold:
                significant.add(field)
        return significant

    uf_fields = _fields(uf_index)
    otel_fields = _fields(otel_index)

    missing = uf_fields - otel_fields
    assert not missing, (
        f"Fields present in UF index '{uf_index}' (sourcetype={sourcetype}) but "
        f"missing from OTel index '{otel_index}': {sorted(missing)}"
    )
    return uf_fields, otel_fields


# ---------------------------------------------------------------------------
# Timestamp delta
# ---------------------------------------------------------------------------

def timestamp_delta(service, uf_index, otel_index, sourcetype,
                    max_avg_delta_seconds=60, earliest="-24h", latest="now"):
    """
    Assert there is no systematic indexing-time vs event-time offset.

    Computes avg(_indextime - _time) per index and checks that neither index
    has an average offset exceeding *max_avg_delta_seconds*.
    """
    def _avg_delta(index):
        row = splunk_lib.search_one(
            service,
            f"search index={index} sourcetype={sourcetype} "
            f"| eval delta=_indextime-_time | stats avg(delta) as avg_delta",
            earliest=earliest,
            latest=latest,
        )
        if not row or row.get("avg_delta") in (None, ""):
            return None
        return float(row["avg_delta"])

    uf_delta = _avg_delta(uf_index)
    otel_delta = _avg_delta(otel_index)

    if uf_delta is None or otel_delta is None:
        return  # not enough data

    assert abs(uf_delta) <= max_avg_delta_seconds, (
        f"UF index '{uf_index}' has systematic timestamp offset for "
        f"sourcetype '{sourcetype}': avg(_indextime-_time)={uf_delta:.1f}s"
    )
    assert abs(otel_delta) <= max_avg_delta_seconds, (
        f"OTel index '{otel_index}' has systematic timestamp offset for "
        f"sourcetype '{sourcetype}': avg(_indextime-_time)={otel_delta:.1f}s"
    )
