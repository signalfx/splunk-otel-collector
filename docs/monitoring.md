# Monitoring

The default configuration automatically scrapes the Collector's own metrics and
sends the data using the `signalfx` exporter. A built-in dashboard provides
information about the health and status of Collector instances.

In addition, logs should be collected. For Log Observer customers, logs are
automatically collected for the Collector and Journald processes.

The Collector also offers
[zpages](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/troubleshooting.md#zpages).
