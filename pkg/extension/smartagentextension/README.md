# Smart Agent Extension

The `smartagent` extension provides a mechanism to specify config options that are not
just specific to a single instance of the [Smart Agent Receiver](../../receiver/smartagentreceiver/README.md) but are applicable to
all instances.  This component provides a means of migrating your existing
[Smart Agent configuration](https://docs.splunk.com/observability/en/gdi/opentelemetry/smart-agent/smart-agent-migration-process.html#locate-your-existing-smart-agent-configuration-file)
to the Splunk Distribution of OpenTelemetry Collector.

As the Smart Agent Receiver doesn't provide 1:1 functional parity with the SignalFx Smart Agent in itself,
only a subset of existing configuration fields are supported by the Smart Agent Extension:

1. The `bundleDir` field refers to the path of a supported Smart Agent release bundle when
you need to provide one explicitly for legacy Smart Agent compatibility.
1. `procPath` for host or mounted container procfs access (default `/proc`)
1. `etcPath` for host or mounted container volume/filesystem etc content (default `/etc`)
1. `varPath` for host or mounted container volume/filesystem var content (default `/var`)
1. `runPath` for host or mounted container volume/filesystem run content (default `/run`)
1. `sysPath` for host or mounted container sysfs access (default `/sys`)

In the below example configuration, `bundleDir` and host filesystem paths will be used for all instances
of the `smartagent` receiver.

```yaml
extensions:
  smartagent:
    bundleDir: /bundle/
    procPath: /custom/proc
```
