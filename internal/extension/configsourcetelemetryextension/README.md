# Config Source Telemetry Extension

| Status                   |                                     |
|--------------------------|-------------------------------------|
| Stability                | [alpha]                             |
| Supported pipeline types | N/A (extension)                     |
| Distributions            | [Splunk OpenTelemetry Collector]    |

This extension bridges the gap between the Splunk collector's custom config source
provider and the OpenTelemetry service telemetry infrastructure. It is injected
automatically into the collector's service configuration at startup — no user
configuration is required.

## Purpose

The Splunk OpenTelemetry Collector supports custom config sources (e.g. `vault`,
`include`, `env`, `zookeeper`, `etcd2`) that are initialised before the collector
service — and therefore before the service's `MeterProvider` — is available.
This extension is started by the service with a fully initialised
`component.TelemetrySettings`. On `Start()` it injects those settings into the
global `TelemetryHook`, enabling the hook to register the
`otelcol_splunk_configsource_usage` observable gauge against the **service's own**
`MeterProvider`. The metric then appears at the collector's internal Prometheus
endpoint (`/metrics`) and is automatically scraped and forwarded to your observability
backend by the existing `prometheus/internal` receiver pipeline.

## Configuration

This extension requires no configuration. It is automatically injected into the
collector's config by the `InjectConfigSourceTelemetryExtension` config converter,
which runs at startup before the service initialises.

If for any reason you need to declare it explicitly, the minimal config is:

```yaml
extensions:
  configsource_telemetry:

service:
  extensions: [configsource_telemetry]
```

## Metrics

| Metric name                         | Type  | Attributes           | Description |
|-------------------------------------|-------|----------------------|-------------|
| `otelcol_splunk_configsource_usage` | Gauge | `config_source_type` | Emitted with value `1` for each custom config source that is present in the collector config. No datapoint is emitted for config sources that are not in use. |

### Attribute values for `config_source_type`

| Value       | Config source           |
|-------------|-------------------------|
| `env`       | `envvarconfigsource`    |
| `include`   | `includeconfigsource`   |
| `vault`     | `vaultconfigsource`     |
| `zookeeper` | `zookeeperconfigsource` |
| `etcd2`     | `etcd2configsource`     |

### Example output

When `vault` and `include` are the active config sources the metric endpoint
exposes exactly two datapoints — one per active source:

```
# HELP otelcol_splunk_configsource_usage Indicates whether a custom config source is in use (1 = in use)
# TYPE otelcol_splunk_configsource_usage gauge
otelcol_splunk_configsource_usage{config_source_type="include"} 1
otelcol_splunk_configsource_usage{config_source_type="vault"} 1
```

Config sources that are not in use produce no datapoint.
