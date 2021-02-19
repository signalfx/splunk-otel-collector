# SignalFx Smart Agent Extension

The `smartagent` extension provides a mechanism to specify config options that are not
just specific to a single instance of the [`smartagent` receiver](../../receiver/smartagentreceiver/README.md) but are applicable to
all instances of the `smartagent` receiver.

To begin with, this extension will provide a mechanism to specify config options to configure
collectd. These options are mapped to the [collectd config options](https://docs.signalfx.com/en/latest/integrations/agent/config-schema.html#collectd)
in the SignalFx Agent. Note that if this extension is not configured, the defaults in `smartagent`
receiver will be used.

In the below example configuration, `configDir` and `bundleDir` will be used for all instances
of the `smartagent` receiver that wrap around a collectd based monitor.

```yaml
extensions:
  smartagent:
    bundleDir: /bundle/
    collectd:
      configDir: /etc/collectd/
```

The full list of settings exposed for this receiver are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).
