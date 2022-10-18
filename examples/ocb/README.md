# Notes

**Upgrading to a new version of the open source release**

1. Update the `checksums.txt` file by directly going to the
github.com [OpenTelemetry Collector Releases
page](https://github.com/open-telemetry/opentelemetry-collector/releases),
and downloading the `checksums.txt` file for the new version into
this repo.

2. Update the Makefile `OTEL_COLLECTOR_VERSION` to the new
release number, e.g. `0.61.0` (without the `v` letter prefix).

3. Commit to git with `git add checksums.txt Makefile` and then `git
commit -m "update OpenTelemetry Collector version"`.

4. Open a PR with the changes.
