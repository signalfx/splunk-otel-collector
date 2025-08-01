# Smart Agent Receiver

This receiver allows to use monitors.

Monitors collect metrics from the host system and services. They are configured under the monitors list in the agent config. For application-specific monitors, you can define discovery rules in your monitor configuration. A separate monitor instance is created for each discovered instance of applications that match a discovery rule. See [Auto Discovery](https://github.com/signalfx/signalfx-agent/blob/main/docs/auto-discovery.md) for more information.

Many of the monitors are built around [collectd](https://collectd.org/), an open source third-party monitor, and use it to collect metrics. Some other monitors do not use collectd. However, either type is configured in the same way.

For a list of supported monitors and their configurations, see [Monitor Config](https://github.com/signalfx/signalfx-agent/blob/main/docs/monitor-config.md).

The agent is primarily intended to monitor services/applications running on the same host as the agent. This is in keeping with the collectd model. The main issue with monitoring services on other hosts is that the host dimension that collectd sets on all metrics will currently get set to the hostname of the machine that the agent is running on. This allows everything to have a consistent host dimension so that metrics can be matched to a specific machine during metric analysis.

See the
[migration guide](https://docs.splunk.com/observability/en/gdi/opentelemetry/smart-agent/smart-agent-migration-to-otel-collector.html)
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
1. To replace or modify metrics, use
[Collector processors](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/README.md).
1. If you have a monitor that sends [events](https://dev.splunk.com/observability/docs/datamodel/custom_events) (e.g. `kubernetes-events`,
`nagios`, `processlist`, and some `telegraf` monitors like `telegraf/exec`), add it to a `logs` pipeline that uses a
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
  smartagent/postgresql:
    type: postgresql
    host: mypostgresinstance
    port: 5432
    dimensionClients:
      - signalfx  # references the SignalFx Exporter configured below
  smartagent/processlist:
    type: processlist

processors:
  resourcedetection:
    detectors:
      - system

exporters:
  signalfx:
    access_token: "${SIGNALFX_ACCESS_TOKEN}"
    realm: us1
  otlphttp:
    traces_endpoint: "https://ingest.${SPLUNK_REALM}.signalfx.com/v2/trace/otlp"
    headers:
      "X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"

service:
  pipelines:
    metrics:
      receivers:
        - smartagent/postgresql
        - smartagent/kafka
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
        - otlp
      processors:
        - resourcedetection
      exporters:
        - otlphttp
```
