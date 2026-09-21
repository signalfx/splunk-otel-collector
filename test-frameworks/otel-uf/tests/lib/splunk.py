"""
Thin wrapper around splunk-sdk for running one-shot searches.
"""
import time
import splunklib.results as results


def search(service, spl, earliest="-24h", latest="now", max_count=50000):
    """
    Run a blocking Splunk search and return a list of result dicts.

    Parameters
    ----------
    service   : splunklib.client.Service
    spl       : str   - SPL query (must start with 'search' or '|')
    earliest  : str   - earliest time bound (default: last 24 hours)
    latest    : str   - latest time bound (default: now)
    max_count : int   - maximum number of results to return

    Returns
    -------
    list[dict]  - one dict per result row
    """
    kwargs = {
        "exec_mode": "blocking",
        "earliest_time": earliest,
        "latest_time": latest,
        "count": max_count,
    }
    job = service.jobs.create(spl, **kwargs)

    reader = results.JSONResultsReader(job.results(output_mode="json", count=max_count))
    rows = [r for r in reader if isinstance(r, dict)]
    job.cancel()
    return rows


def search_one(service, spl, **kwargs):
    """Return the first result row, or None if the search returned no rows."""
    rows = search(service, spl, **kwargs)
    return rows[0] if rows else None
