# Fluentd support removal

Fluentd support has been deprecated, and in some cases, removed.

## Replacement guidance

Please use native OTel Collector receivers instead. A common replacement for Fluentd's functionality is the
[filelog receiver](https://help.splunk.com/en/splunk-observability-cloud/manage-data/available-data-sources/supported-integrations-in-splunk-observability-cloud/opentelemetry-receivers/filelog-receiver).
Many common configuration examples of the filelog receiver can be found in the [logs_config_linux.yaml](https://github.com/signalfx/splunk-otel-collector/blob/87bee7ae45b08be8d143a758d0f7004fd92d8f60/cmd/otelcol/config/collector/logs_config_linux.yaml)
file.