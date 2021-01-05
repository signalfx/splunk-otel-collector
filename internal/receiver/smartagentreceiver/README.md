# SignalFx Smart Agent Receiver

The Smart Agent Receiver allows you to utilize existing [SignalFx Smart Agent monitors](https://github.com/signalfx/signalfx-agent#monitors)
as OpenTelemetry Collector metric receivers.  It assumes that you have a properly configured environment with a
functional [Smart Agent release bundle](https://github.com/signalfx/signalfx-agent/releases/latest) on your system.

**pre-alpha: Not intended to be used and no stability or functional guarantees are made of any kind at this time.**

## Configuration

Each `smartagent` receiver configuration acts a drop-in replacement for each supported Smart Agent Monitor
[configuration](https://github.com/signalfx/signalfx-agent/blob/master/docs/monitor-config.md) with some exceptions:

1. In lieu of `discoveryRule` support, the Collector's
[`receivercreator`](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/master/receiver/receivercreator/README.md)
and associated [Observer extensions](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/extension/observer/README.md)
should be used.
1. All metric content replacement and transformation rules should utilize existing
[Collector processors](https://github.com/open-telemetry/opentelemetry-collector/blob/master/processor/README.md).

Example:

```yaml
receivers:
  smartagent/haproxy:
    type: haproxy
    host: myhaproxyinstance
    port: 8080
  smartagent/postgresql:
    type: postgresql
    host: mypostgresinstance
    port: 5432
```

The full list of settings exposed for this receiver are documented for
[each monitor](https://github.com/signalfx/signalfx-agent/tree/master/docs/monitors), and the implementation is
[here](./config.go)
