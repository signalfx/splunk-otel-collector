# SignalFx Smart Agent Extension

The `smartagent` extension provides a mechanism to specify config options that are not
just specific to a single instance of the [Smart Agent Receiver](../../receiver/smartagentreceiver/README.md) but are applicable to
all instances.  This component provides a means of migrating your existing
[Smart Agent configuration](https://docs.signalfx.com/en/latest/integrations/agent/config-schema.html#config-schema)
to the Splunk Distribution of OpenTelemetry Collector.

As the Smart Agent Receiver doesn't provide 1:1 functional parity with the SignalFx Smart Agent in itself,
only a subset of existing configuration fields are supported by the Smart Agent Extension at this time:

1. The [bundleDir] field refers to the path of a supported Smart Agent release bundle.  All provided
Splunk Distribution of OpenTelemetry Collector packages include the agent bundle, and their installers
source its value via the `SPLUNK_BUNDLE_DIR` environment variable by default.
1. The [collectd](https://docs.signalfx.com/en/latest/integrations/agent/config-schema.html#collectd)
field refers to performance and debugging configurables for the collectd subprocess and associated mechanisms.
If the Smart Agent Extension or this field are not configured, the Agent defaults will be inherited.
This configuration object's `configDir` refers to the location for internal configuration files and is set to the value
of the `SPLUNK_COLLECTD_DIR` environment variable by the default agent deployment mode config.
1. `procPath` for host or mounted container procfs access (default `/proc`)
1. `etcPath` for host or mounted container volume/filesystem etc content (default `/etc`)
1. `varPath` for host or mounted container volume/filesystem var content (default `/var`)
1. `runPath` for host or mounted container volume/filesystem run content (default `/run`)
1. `sysPath` for host or mounted container sysfs access (default `/sys`)

In the below example configuration, `configDir` and `bundleDir` will be used for all instances
of the `smartagent` receiver that wrap around a collectd based monitor.

```yaml
extensions:
  smartagent:
    bundleDir: /bundle/
    procPath: /custom/proc
    collectd:
      configDir: /tmp/collectd/config
```
