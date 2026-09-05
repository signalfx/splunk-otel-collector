# PromQL Receiver
| Status        |           |
| ------------- |-----------|
| Stability     | [development]: metrics   |
| Distributions | [splunk] |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development
[splunk]: https://github.com/signalfx/splunk-otel-collector

# Configuration

| Option | Default | Description |
| ------ | ------- | ----------- |
| `endpoint` | (none, required) | Base URL of the Prometheus HTTP API (e.g. `http://localhost:9090`) to send queries against. |
| `queries` | (none, required) | List of PromQL queries to execute. Must contain at least one entry. |
| `queries[].query` | (none, required) | The PromQL expression to run against the endpoint. |
| `queries[].metric_name_fallback` | (none) | Metric name to use when the query result has no `__name__` label (e.g. after an aggregation like `sum()`). Falls back to `promql_result` if unset. |
| `collection_interval` | `1m` | How often the queries are executed. |
| `initial_delay` | `1s` | Delay before the first collection after the receiver starts. |
| `timeout` | (none) | Timeout applied to each collection interval. |
| `tls` | (none) | TLS client settings used when connecting to the endpoint, see [confighttp](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/README.md). |


