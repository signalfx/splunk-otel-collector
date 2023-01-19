# SignalFx Smart Agent Configuration Translation Tool

This package provides a command-line tool, `translatesfx`, that translates a 
SignalFx Smart Agent configuration file into a configuration that can be
used by a Splunk Distribution of OpenTelemetry Collector. The `translatesfx`
tool is intended to be part of a larger migration process covered in the document
[Migrating from SignalFx Smart Agent to Splunk Distribution of OpenTelemetry Collector](https://docs.splunk.com/Observability/gdi/opentelemetry/smart-agent-migration-to-otel-collector.html).

## Caveats

This tool has been used against several example configs and has been shown 
in some cases to produce OpenTelemetry Collector configs that are fully 
functional and comparable to the original Smart Agent config. However, 
this tool is designed to produce only a reasonably accurate approximation of 
the final, production OTel config that would replace a Smart Agent config. It is 
not designed to, and cannot in all cases, produce a drop-in replacement OTel
config for a Smart Agent one.

Instead, this tool aims to automate most of the config changes required when migrating
to OTel from Smart Agent; any config produced by this tool should be carefully 
evaluated and tested before being put into production.

## Where to Get It

`translatesfx` executables are available on the
[releases page](https://github.com/signalfx/splunk-OTel-collector/releases).
and are contained in the rpm, msi, and deb packages as well as the docker images
(v0.36.1 and up).

## Usage

The `translatesfx` command requires one argument, a Smart Agent 
configuration file, and accepts an optional second argument, the working directory 
used by any Smart Agent `#from` file expansion directives. The `translatesfx` command
uses this working directory to resolve any relative paths to files referenced by 
any `#from` directives at runtime. If this working directory argument is omitted, `translatesfx`
expands relative file paths using the current working directory.

```
% translatesfx <sfx-file> [<file expansion working directory>]
```

When `translatesfx` runs, it sends the translated OpenTelemetry Collector configuration
yaml to standard output. To write the contents to disk, you could redirect this output
to a new OTel configuration file:

```
% translatesfx sa-config.yaml > otel-config.yaml
```

## Examples

#### CLI Usage

###### Using the current working directory to expand files

```
% translatesfx path/to/sfx/config.yaml
```

###### Using a custom working directory to expand files

```
% translatesfx path/to/sfx/config.yaml path/to/sfx
```

#### Basic example

Given the following input
```yaml
signalFxAccessToken: {"#from": "env:SFX_ACCESS_TOKEN"}
ingestUrl: {"#from": "ingest_url", default: "https://ingest.signalfx.com"}
apiUrl: {"#from": "api_url", default: "https://api.signalfx.com"}
traceEndpointUrl: {"#from": 'trace_endpoint_url', default: "https://ingest.signalfx.com/v2/trace"}

intervalSeconds: 10

logging:
  level: info

monitors:
  - {"#from": "monitors/*.yaml", flatten: true, optional: true}
  - type: memory
```

and the following included files:

###### ingest_url
```
https://ingest.us1.signalfx.com
```

###### api_url
```
https://api.us1.signalfx.com
```

###### trace_endpoint_url
```
https://ingest.signalfx.com/v2/trace
```

###### monitors/cpu.yaml
```yaml
- type: cpu
```

###### monitors/load.yaml
```yaml
- type: load
```

`translatesfx` would produce:
```yaml
receivers:
  smartagent/cpu:
    type: cpu
  smartagent/load:
    type: load
  smartagent/memory:
    type: memory
exporters:
  signalfx:
    access_token: ${SFX_ACCESS_TOKEN}
    realm: us1
service:
  pipelines:
    metrics:
      receivers:
      - smartagent/cpu
      - smartagent/load
      - smartagent/memory
      exporters:
      - signalfx
```

## Translatable features

#### Monitors

Smart Agent monitors are translated into OTel receivers and placed into the
appropriate pipelines.

For example, the Smart Agent monitor:

```yaml
monitors:
  - type: vsphere
    host: 1.2.3.4
    username: user
    password: abc123
```

would be translated into an OTel receiver and placed into the `metrics` pipeline:

```yaml
receivers:
  smartagent/vsphere:
    type: vsphere
    host: 1.2.3.4
    username: user
    password: abc123
service:
  pipelines:
    metrics:
      receivers:
        - smartagent/vsphere
```

#### Config Sources

Smart Agent
[remote configuration](https://github.com/signalfx/signalfx-agent/blob/main/docs/remote-config.md)
directives are translated into their corresponding OTel config sources.

The types of config sources supported for translation are the following:

###### Environment variables

Smart Agent environment variable interpolation in the form of `${VARNAME}` works similarly in OTel,
so these are simply retained as is.

Also supported is the more advanced form:

```yaml
signalFxAccessToken: {"#from": "env:SIGNALFX_ACCESS_TOKEN"}
```

which would be translated into:

```yaml
${SIGNALFX_ACCESS_TOKEN}
```

**Note:** Support for default environment variable values is not available, but may be added in a future release.

Docs:
[Smart Agent](https://github.com/signalfx/signalfx-agent/blob/main/docs/remote-config.md#environment-variables)
|
[OTel](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/envvarconfigsource)

###### Include

Smart Agent file include directives are translated into OTel include config-source directives.

For example, the Smart Agent file include:

```yaml
signalFxAccessToken: {"#from": "/etc/signalfx/token"}
```

would be translated into an OTel `include` config-source:

```yaml
config_sources:
  include:
exporters:
  signalfx:
    access_token: "${include:/etc/signalfx/token}"
```

Additionally, globbed paths, `flatten`, and `default` values are supported but only at `translatesfx` tool run time.
This is because OTel's config-source functionality generally doesn't support these features. As a result, a directive with these
attributes will be attempted to be expanded/inlined when the tool is run if the referenced files are available. For this reason,
it is also recommended that this tool be run in the same environment where the Smart Agent runs -- that way it can read
any referenced files if it needs to.

Note: support for disabling inlining may be added in a future release.

Docs:
[Smart Agent](https://github.com/signalfx/signalfx-agent/blob/main/docs/remote-config.md)
|
[OTel](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/includeconfigsource)

###### Zookeeper

Smart Agent Zookeeper remote configs are translated into OTel zookeeper config-source directives.

For example, the Smart Agent zookeeper config source:

```yaml
configSources:
  zookeeper:
    endpoints:
      - 127.0.0.1:2181
    timeoutSeconds: 10
monitors:
  - type: collectd/redis
    host: localhost
    port: 1234
    auth: {"#from": "zookeeper:/redis/password"}
```

would be translated into an OTel config source:

```yaml
config_sources:
  zookeeper:
    endpoints:
    - 127.0.0.1:2181
    timeout: 10s
receivers:
  smartagent/collectd/redis:
    type: collectd/redis
    host: localhost
    port: 1234
    auth: ${zookeeper:/redis/password}
```

Docs:
[Smart Agent](https://github.com/signalfx/signalfx-agent/blob/main/docs/config-schema.md#zookeeper)
|
[OTel](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/zookeeperconfigsource)

###### Etcd2

Smart Agent etcd2 config-source directives are translated into their OTel equivalent.

For example, the following Smart Agent config-source:

```yaml
configSources:
  etcd2:
    endpoints:
      - http://127.0.0.1:2379
    username: foo
    password: bar
monitors:
  - type: collectd/redis
    host: localhost
    port: 1234
    auth: {"#from": "etcd2:/redispassword"}
```

would be translated into an OTel config-source:

```yaml
config_sources:
  etcd2:
    auth:
      password: bar
      username: foo
    endpoints:
    - http://127.0.0.1:2379
receivers:
  smartagent/collectd/redis:
    type: collectd/redis
    host: localhost
    port: 1234
    auth: ${etcd2:/redispassword}

```

Docs:
[Smart Agent](https://github.com/signalfx/signalfx-agent/blob/main/docs/config-schema.md#etcd2)
|
[OTel](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/etcd2configsource)

###### Vault

The Smart Agent Vault config-source:

```yaml
configSources:
  vault:
    vaultAddr: http://127.0.0.1:8200
    vaultToken: abc123
monitors:
  - type: collectd/redis
    host: localhost
    port: 1234
    auth: {"#from": "vault:/secret/redis[password]"}
```

would be translated into an OTel equivalent:

```yaml
config_sources:
  vault/0:
    endpoint: http://127.0.0.1:8200
    path: /secret/redis
    auth:
      token: abc123
receivers:
  smartagent/collectd/redis:
    type: collectd/redis
    host: localhost
    port: 1234
    password: ${vault/0:password}
```

Docs:
[Smart Agent](https://github.com/signalfx/signalfx-agent/blob/main/docs/config-schema.md#vault)
|
[OTel](https://github.com/signalfx/splunk-otel-collector/tree/main/internal/configsource/vaultconfigsource)

#### Global Dimensions

Smart Agent `globalDimensions` are translated into an OTel metrics transform processor placed into the generated metrics pipeline.

For example, the following Smart Agent global dimensions:

```yaml
globalDimensions:
  aaa: 42
  bbb: 111
```

would be translated into the following OTel metrics transform processor placed into the generated metrics pipeline:

```yaml
processors:
  metricstransform:
    transforms:
      - action: update
        include: .*
        match_type: regexp
        operations:
          - action: add_label
            new_label: aaa
            new_value: 42
          - action: add_label
            new_label: bbb
            new_value: 111
```

#### Resource Detection

The Smart Agent performs cloud resource detection out of the box. To 
provide similar functionality, `translatesfx` creates a resource
detection processor with a list of cloud detectors. Like Smart Agent,
these set the hostname and other dimensions, which are then used
by the `signalfx` exporter to generate cloud resource IDs.

```yaml
  resourcedetection:
    detectors:
      - env
      - gcp
      - ecs
      - ec2
      - azure
      - system
```

#### Metrics To Include/Exclude

Smart Agent `metricsToExclude` are translated into a `filter` processor
placed into the generated metrics pipeline.

For example, the Smart Agent metrics to exclude:

```yaml
metricsToExclude:
  - metricNames:
      - node_filesystem_*
      - '!node_filesystem_free_bytes'
```

would be translated into the following filter processor, placed into the generated
metrics pipeline:

```yaml
processors:
  filter:
    metrics:
      exclude:
        match_type: expr
        expressions:
          - MetricName matches "^node_filesystem_.*$" and not (MetricName matches "^node_filesystem_free_bytes$")
            and (not (MetricName matches "^node_filesystem_readonly$"))
```

#### Discovery Rules

Smart Agent `discoveryRule`s work with `observers` to dynamically configure
and start monitors. 

The corresponding functionality in OTel is handled by `receiver_creator` and `watch_observers`.
The translation tool translates observers and discovery rules to a receiver creator with OTel rules and watch observers,
replacing Smart Agent identifiers and operators found in discovery rules with corresponding OTel ones.

For example, the Smart Agent observer and discovery rule:

```yaml
observers:
  - type: k8s-api
monitors:
  - type: collectd/redis
    discoveryRule: kubernetes_pod_name == "redis" && port == 6379
```

would be translated into an OTel observer, rule, and receiver creator:

```yaml
extensions:
  k8s_observer:
    auth_type: serviceAccount
    node: ${K8S_NODE_NAME}
receivers:
  receiver_creator:
    receivers:
      smartagent/collectd/redis:
        config:
          type: collectd/redis
        rule: type == "hostport" && process_name matches "redis" && port == 6379
    watch_observers:
      - k8s_observer
```

#### Special Cases

* If a `processlist` monitor is found in the SA config, a corresponding receiver placed into a `logs` pipeline in the OTel config 
* If a `signalfx-forwarder` monitor is found in the SA config, a corresponding receiver is placed into a `traces` pipeline, containing both a `sapm` and a `signalfx` exporter
* All pipelines get a `resourcedetection` processor containing the default list of cloud detectors
