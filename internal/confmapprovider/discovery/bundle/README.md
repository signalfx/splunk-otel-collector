## bundle.d

`bundle.d` refers to the [`embed.FS`](https://pkg.go.dev/embed#hdr-File_Systems) config directory made available by the
[`bundle.BundledFS`](./bundle.go). It currently consists of all `./bundle.d/extensions/*.discovery.yaml` and
`./bundle.d/receivers/*.discovery.yaml` files that are generated by the `discoverybundler` cmd.

To construct the latest bundle.d contents before building the collector run:

```bash
$ make bundle.d
```

### *.discovery.yaml.tmpl

All discovery config component discovery.yaml files are generated from [`text/template`](https://pkg.go.dev/text/template)
`discovery.yaml.tmpl` files using built-in validators and property guidance helpers:

Example `redis.discovery.yaml.tmpl`:

```yaml
{{ receiver "redis" }}:
  enabled: true
  rule:
    docker_observer: type == "container" and port == 6379
    <...>
  status:
    <...>
    statements:
      partial:
        - regexp: 'ERR AUTH.*'
          message: >-
            Please ensure your redis password is correctly specified with
            `{{ configPropertyEnvVar "password" "<username>" }}` environment variable.
```

After adding the required new component filename prefix to the `Components` instance in [`components.go`](./components.go)
and running `make bundle.d`, there's now a corresponding `bundle.d/receivers/redis.discovery.yaml`:

```yaml
#####################################################################################
# This file is generated by the Splunk Distribution of the OpenTelemetry Collector. #
#####################################################################################
redis:
  enabled: true
  rule:
    docker_observer: type == "container" and port == 6379
  <...>
  status:
  <...>
    statements:
      partial:
        - regexp: 'ERR AUTH.*'
          message: >-
            Please ensure your redis password is correctly specified with
            `--set splunk.discovery.receivers.redis.config.password="<password>"` or
            `SPLUNK_DISCOVERY_RECEIVERS_redis_CONFIG_password="<username>"` environment variable.
```

In order for this to be included in the [Windows](./bundledfs_windows.go) and [Linux](./bundledfs_others.go) `BundledFS`
be sure to include the component filename prefix in the corresponding `Components.Linux` and `Components.Windows`
functions.

When building the collector afterward, this redis receiver discovery config is now made available to discovery mode, and
it can be disabled by `--set splunk.discovery.receivers.redis.enabled=false` or
`SPLUNK_DISCOVERY_RECEIVERS_redis_ENABLED=false`.
