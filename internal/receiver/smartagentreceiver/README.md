# SignalFx Smart Agent Receiver

The Smart Agent Receiver allows you to utilize existing [SignalFx Smart Agent monitors](https://github.com/signalfx/signalfx-agent#monitors)
as OpenTelemetry Collector metric receivers.  It assumes that you have a properly configured environment with a
functional [Smart Agent release bundle](https://github.com/signalfx/signalfx-agent/releases/latest) on your system.

**Beta: No stability or functional guarantees are made at this time.  Configuration and behavior may change without notice.**

## Configuration

Each `smartagent` receiver configuration acts a drop-in replacement for each supported Smart Agent Monitor
[configuration](https://github.com/signalfx/signalfx-agent/blob/master/docs/monitor-config.md) with some exceptions:

1. In lieu of `discoveryRule` support, the Collector's
[`receivercreator`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/receiver/receivercreator/README.md)
and associated [Observer extensions](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/extension/observer/README.md)
should be used.
1. All metric content replacement and transformation rules should utilize existing
[Collector processors](https://github.com/open-telemetry/opentelemetry-collector/blob/master/processor/README.md).
1. Monitors with [dimension property and tag update
functionality](https://dev.splunk.com/observability/docs/datamodel#Creating-or-updating-custom-properties-and-tags)
allow an associated `dimensionClients` field that references the name of the SignalFx exporter you are using in your
pipeline.  These monitors include `ecs-metadata`, `heroku-metadata`, `kubernetes-cluster`, `openshift-cluster`, `postgresql`,
and `sql`.
If you do not specify any exporters via this field, the receiver will attempt to use the associated
pipeline.  If the next element of the pipeline isn't compatible with dimension update behavior and if a lone SignalFx
exporter was configured for your deployment, it will be selected.  If no dimension update behavior is desired,
you can specify the empty array `[]` to disable.
1. Monitors with [event-sending
functionality](https://dev.splunk.com/observability/docs/datamodel/ingest#Send-custom-events) should also be made members of
a `logs` pipeline that utilizes a [SignalFx
exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/signalfxexporter/README.md)
that will make the event submission requests.  It's recommended, and in the case of the Processlist monitor required,
to use a Resource Detection to ensure that host identity and other useful information is made available as event
dimensions.  Receiver entries that should be added to logs pipelines `nagios`, `processlist`, and potentially any
`telegraf/*` monitors like `telegraf/exec`.

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
```

For a more detailed description of migrating your Smart Agent monitor usage to the Splunk distribution of
OpenTelemetry Collector please see the [migration guide](../../../docs/signalfx-smart-agent-migration.md).
