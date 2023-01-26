# SignalFx Smart Agent Receiver

> **Note:** The SignalFx Smart Agent receiver is fully supported only on x86_64/amd64 platforms.

The Smart Agent Receiver lets you use [SignalFx Smart Agent monitors](https://github.com/signalfx/signalfx-agent#monitors)
in the [Splunk Distribution of OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector). Many
monitors also require a [Smart Agent release bundle](https://github.com/signalfx/signalfx-agent/releases/latest),
which is installed by the Splunk Distribution of OpenTelemetry Collector on supported x86_64/amd64 platforms.

See the
[migration guide](../../../docs/signalfx-smart-agent-migration.md)
for more information about migrating from the Smart Agent to the Splunk Distribution of the OpenTelemetry Collector.

**Beta: All Smart Agent monitors are supported by Splunk. Configuration and behavior may change without notice.**

## Configuration

For each Smart Agent 
[monitor](https://github.com/signalfx/signalfx-agent/blob/main/docs/monitor-config.md)
you want to add to the Collector, add a `smartagent` receiver configuration block. Once configured in the Collector, each
`smartagent` receiver acts as a drop-in replacement for its corresponding Smart Agent monitor.

1. Put any Smart Agent or collectd configuration into the global
[Smart Agent Extension section](../../extension/smartagentextension/README.md)
of your Collector configuration.
1. Instead of using `discoveryRule`, use the Collector's
[Receiver Creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/receivercreator/README.md)
and [Observer extensions](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/README.md).
1. If you're using a [SignalFx Forwarder](https://github.com/signalfx/signalfx-agent/blob/main/docs/monitors/signalfx-forwarder.md)
monitor, put it into both a `traces` and a `metrics` pipeline, and use a
[Sapm exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/sapmexporter/README.md)
and a 
[SignalFx exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/signalfxexporter/README.md),
as each pipeline's exporter, respectively.
1. To replace or modify metrics, use
[Collector processors](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/README.md).
1. If you have a monitor that sends [events](https://dev.splunk.com/observability/docs/datamodel/custom_events) (e.g. `kubernetes-events`,
`nagios`, `processlist`, and some `telegraf` monitors like `telegraf/exec`), put it in a `logs` pipeline that uses a
[SignalFx exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/signalfxexporter/README.md).
It's recommended, and in the case of the Processlist monitor required, to put into the same pipeline a
[Resource Detection processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/resourcedetectionprocessor/README.md),
which will add host information and other useful dimensions to the events. An example is provided below.
1. If you have a monitor that updates [dimension properties or tags](https://dev.splunk.com/observability/docs/datamodel/metrics_metadata), for example `ecs-metadata`, `heroku-metadata`, `kubernetes-cluster`, `openshift-cluster`, `postgresql`, or `sql`, put the name of
your SignalFx exporter in its `dimensionClients` field in the Collector's SignalFx receiver configuration block.
If you don't specify any exporters in this array field, the receiver attempts to use the Collector pipeline to which it's connected. If
the next element of the pipeline isn't compatible with updating dimensions, and if you configured a single SignalFx exporter,
the receiver uses that SignalFx exporter. If you don't require dimension updates, you can specify the empty array `[]` to disable it.


Example:

```yaml
receivers:
  smartagent/signalfx-forwarder:
    type: signalfx-forwarder
  smartagent/postgresql:
    type: postgresql
    host: mypostgresinstance
    port: 5432
    dimensionClients:
      - signalfx  # references the SignalFx Exporter configured below
  smartagent/processlist:
    type: processlist
  smartagent/kafka:
    type: collectd/kafka
    host: mykafkabroker
    port: 7099
    clusterName: mykafkacluster
    intervalSeconds: 5

processors:
  resourcedetection:
    detectors:
      - system

exporters:
  signalfx:
    access_token: "${SIGNALFX_ACCESS_TOKEN}"
    realm: us1
  sapm:
    access_token: "${SIGNALFX_ACCESS_TOKEN}"
    endpoint: https://ingest.us1.signalfx.com/v2/trace

service:
  pipelines:
    metrics:
      receivers:
        - smartagent/postgresql
        - smartagent/kafka
        - smartagent/signalfx-forwarder
      processors:
        - resourcedetection
      exporters:
        - signalfx
    logs:
      receivers:
        - smartagent/processlist
      processors:
        - resourcedetection
      exporters:
        - signalfx
    traces:
      receivers:
        - smartagent/signalfx-forwarder
      processors:
        - resourcedetection
      exporters:
        - sapm
```
