# Oracle Encoding Extension

| Status        |             |
|---------------|-------------|
| Stability     | [alpha]     |
| Distributions | [Splunk OpenTelemetry Collector] |

The `oracle_encoding` extension unmarshals metrics published by
[OCI (Oracle Cloud Infrastructure) Monitoring](https://docs.oracle.com/en-us/iaas/Content/Monitoring/Concepts/monitoringoverview.htm)
in JSONL format (one JSON object per line) into OpenTelemetry metrics. It
implements the
[`encoding.MetricsUnmarshalerExtension`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/encoding/encoding.go)
interface, so it can be referenced from the `encoding` config option of any
receiver that supports pluggable encoding extensions (e.g. `kafkareceiver`).

Only metrics are supported; logs and traces are out of scope for this
extension.

## Configuration

This extension currently requires no configuration.

```yaml
extensions:
  oracle_encoding:
```

## Input format

Each line of input is expected to be a JSON object shaped as follows:

```json
{
  "namespace": "myFirstNamespace",
  "compartmentId": "ocid1.compartment.oc1..exampleuniqueID",
  "resourceGroup": "myFirstResourceGroup",
  "name": "successRate",
  "dimensions": {
    "resourceId": "ocid1.exampleresource.region1.phx.exampleuniqueID",
    "appName": "myAppA"
  },
  "metadata": {
    "unit": "percent",
    "displayName": "MyAppA Success Rate"
  },
  "datapoints": [
    { "timestamp": 1768083560000, "value": 83.0 },
    { "timestamp": 1768083580000, "value": 90.1 }
  ]
}
```

`timestamp` is epoch milliseconds. The `resourceId` dimension may also be
spelled `resourceID`; both are matched case-insensitively.

Records sharing the same `namespace`, `compartmentId`, `resourceGroup` and
`resourceId` dimension are grouped into a single `ResourceMetrics`. These
fields are mapped onto the resource following the
[cloud semantic conventions](https://opentelemetry.io/docs/specs/semconv/registry/attributes/cloud/)
where a generic attribute exists, and the
[`oracle_cloud.*`](https://opentelemetry.io/docs/specs/semconv/registry/attributes/oracle-cloud/)
vendor namespace otherwise:

| OCI field                               | Resource attribute                |
|-----------------------------------------|-----------------------------------|
| (constant)                              | `cloud.provider` = `oracle_cloud` |
| `dimensions.resourceId`                 | `cloud.resource_id`               |
| `compartmentId`                         | `oracle_cloud.compartment_id`     |
| `namespace`                             | `oracle_cloud.namespace`          |
| (computed based on the `compartmentId`) | `oracle_cloud.realm`              |
| `resourceGroup`                         | `oracle_cloud.resource_group`     |

Each record's `datapoints` become a metric named after the record's `name`,
with the remaining `dimensions` applied as datapoint attributes, and
`metadata.unit`/`metadata.displayName` mapped to the metric's
unit/description.

OCI Monitoring does not report an explicit metric type, and `metadata.unit`
is descriptive metadata rather than a signal of temporality or additivity, so
every datapoint is represented as a `Gauge`.

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[Splunk OpenTelemetry Collector]: https://github.com/signalfx/splunk-otel-collector
