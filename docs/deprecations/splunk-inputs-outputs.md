# `splunk_inputs` and `splunk_outputs` removal

The `splunk_inputs` receiver and `splunk_outputs` exporter have been removed. Reading a Splunk
`.conf` tree is now the job of the `splunkconf` config provider, which resolves the same
`inputs.conf`, `outputs.conf`, `props.conf` and `transforms.conf` files using the same Splunk
conf precedence.

The capability is unchanged: the collector still runs a Splunk Technology Add-on (TA) without a
real `splunkd`. What changes is that the `.conf` tree is now read when the collector resolves its
configuration, instead of by a receiver and exporter configured in YAML.

## Replacement guidance

Point `--config` at a `splunkconf://` URI instead of writing a YAML config.

Before:

```yaml
# config.yaml
receivers:
  splunk_inputs:
    base_dir: /var/splunk_home

exporters:
  splunk_outputs:
    base_dir: /var/splunk_home

service:
  pipelines:
    logs:
      receivers: [splunk_inputs]
      exporters: [splunk_outputs]
```

```
otelcol --config=config.yaml --feature-gates=+enableTARunner
```

After:

```
otelcol --config=splunkconf:///var/splunk_home/etc --feature-gates=+enableTARunner
```

The YAML file is no longer needed. The `enableTARunner` feature gate is still required, because
the components the provider emits are registered only under that gate.

### URI form

```
splunkconf://<config-root>?pipeline=<name>
```

`<config-root>` is the directory holding `system/` and `apps/*/`, which is the `etc` directory of
the tree the removed components took as `base_dir`. A `base_dir` of `/var/splunk_home` therefore
becomes a config root of `/var/splunk_home/etc`. The path is read verbatim in any format the OS
supports:

```
splunkconf://etc                              relative path
splunkconf:///opt/splunk/etc                  absolute path
splunkconf://C:\Program Files\Splunk\etc      absolute path with drive letter
```

The provider resolves no `$SPLUNK_HOME` or `$SPLUNK_ETC` and knows nothing about install layouts,
so the caller passes a plain path and the same URI always resolves the same way.

`pipeline` names the emitted pipeline, `logs/<name>`, and defaults to `uf`. To stop reading the
`.conf` tree, omit the `--config=splunkconf://...` argument.

### Directory layout

The expected tree is unchanged:

```
<config-root>/
  apps/
    <TA_name>/
      default/
        inputs.conf
        transforms.conf   # optional
        props.conf        # optional
      local/              # optional, overrides default/
        inputs.conf
  system/
    default/              # built-in system-level defaults
    local/                # optional, site-wide overrides for all TAs
      outputs.conf
```

Conf files are layered in Splunk precedence order: `system/default` < `apps/<TA>/default` <
`apps/<TA>/local` < `system/local`. Multiple TAs can live under `apps/` at once and all of them
are discovered.

## Reload

The removed components watched the `.conf` tree with `fsnotify` and reloaded themselves when a
file changed. The provider does not watch files. Reload now rides the collector's own trigger:
send `SIGHUP`, or push a new configuration over OpAMP, and the provider re-reads the tree. This
matches the Universal Forwarder's own triggered model.

Two upstream feature gates control how much of the collector a reload rebuilds:

- `service.partialReloadReceivers` rebuilds only the receivers when a change is receiver-only.
  Enabled by default.
- `service.partialReload` rebuilds only the affected components more generally. Alpha, disabled
  by default.

With neither enabled, a `SIGHUP` restarts the whole service, which is also the fallback when a
partial reload cannot be applied.

## Known differences

- A configuration that fanned out to an extra exporter alongside `splunk_outputs`, such as
  `debug`, cannot be expressed by layering another `--config` source on top of the provider. The
  provider emits a complete pipeline whose exporter IDs it generates, and a later source would
  have to restate that list to add to it.
