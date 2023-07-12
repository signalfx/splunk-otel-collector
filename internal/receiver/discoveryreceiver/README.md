# Discovery Receiver

| Status                   |                  |
|--------------------------|------------------|
| Stability                | [in-development] |
| Supported pipeline types | logs             |
| Distributions            | [Splunk]         |

The Discovery receiver is a receiver compatible with logs pipelines that allows you to test the functional
status of any receiver type whose target is reported by an Observer. It provides configurable `status`
match rules that evaluate the generated receiver's emitted metrics (if any), or component-level log statements
via the instance's [zap.Logger](https://pkg.go.dev/go.uber.org/zap). It works similarly to the
[Receiver Creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/receivercreator/README.md)
(it actually wraps an internal instance of one), but the resulting dynamically-instantiated receivers don't actually
report their metric content to your metrics pipelines. Instead, the metrics are intercepted by an internal metrics
consumer capable of translating desired metrics to log records based on the `status: metrics` rules you define. All
component-level log statements are similarly intercepted by a log evaluator, and can be translated to emitted log
records based on the `status: statements` rules you define.

The receiver also allows you to emit log records for all
[Endpoint](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/observer/endpoints.go)
events from the specified `watch_observers`. This way you can report your environment as observed by platform-specific
observers in real time, with or without discovering receiver statuses:

```yaml
extensions:
  docker_observer:
receivers:
  discovery:
    watch_observers: [docker_observer]
    log_endpoints: true
exporters:
  logging:
    verbosity: detailed
service:
  extensions: [docker_observer]
  pipelines:
    logs:
      receivers: [discovery]
      exporters: [logging]
```

```
2022-07-27T19:00:32.305Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "logging", "#logs": 3}
2022-07-27T19:00:32.306Z	info	ResourceLog #0
Resource SchemaURL:
Resource labels:
     -> event.type: STRING(endpoint.added)
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
LogRecord #0
ObservedTimestamp: 2022-07-27 19:00:32.30572809 +0000 UTC
Timestamp: 2022-07-27 19:00:32.305727862 +0000 UTC
Severity: info
Body: {"alternate_port":5432,"command":"postgres -c shared_preload_libraries=pg_stat_statements -c wal_level=logical -c max_replication_slots=2","container_id":"58e71612910cd3e2a89d809f1b53fee779c4c81d9bbf36ef37f4c05b04037353","endpoint":"172.17.0.4:5432","host":"172.17.0.4","id":"58e71612910cd3e2a89d809f1b53fee779c4c81d9bbf36ef37f4c05b04037353:5432","image":"postgres","labels":{},"name":"naughty_heisenberg","port":5432,"tag":"latest","transport":"TCP","type":"container"}
Attributes:
     -> type: STRING(endpoint)
     -> image: STRING(postgres)
     -> tag: STRING(latest)
     -> port: STRING(5432)
     -> alternate_port: STRING(5432)
     -> host: STRING(172.17.0.4)
     -> transport: STRING(TCP)
     -> labels: STRING(map[])
     -> endpoint: STRING(172.17.0.4:5432)
     -> name: STRING(naughty_heisenberg)
     -> command: STRING(postgres -c shared_preload_libraries=pg_stat_statements -c wal_level=logical -c max_replication_slots=2)
     -> container_id: STRING(58e71612910cd3e2a89d809f1b53fee779c4c81d9bbf36ef37f4c05b04037353)
     -> id: STRING(58e71612910cd3e2a89d809f1b53fee779c4c81d9bbf36ef37f4c05b04037353:5432)
Trace ID:
Span ID:
Flags: 0
LogRecord #1
ObservedTimestamp: 2022-07-27 19:00:32.305771365 +0000 UTC
Timestamp: 2022-07-27 19:00:32.305771316 +0000 UTC
Severity: info
Body: {"alternate_port":0,"command":"nginx -g daemon off;","container_id":"2567fdbc764706d29120b01efcc3a310d87e9e121ec9debbc977d66f5497cdda","endpoint":"172.17.0.2:80","host":"172.17.0.2","id":"2567fdbc764706d29120b01efcc3a310d87e9e121ec9debbc977d66f5497cdda:80","image":"nginx","labels":{"maintainer":"NGINX Docker Maintainers \u003cdocker-maint@nginx.com\u003e"},"name":"ecstatic_davinci","port":80,"tag":"latest","transport":"TCP","type":"container"}
Attributes:
     -> type: STRING(endpoint)
     -> labels: STRING(map[maintainer:NGINX Docker Maintainers <docker-maint@nginx.com>])
     -> id: STRING(2567fdbc764706d29120b01efcc3a310d87e9e121ec9debbc977d66f5497cdda:80)
     -> tag: STRING(latest)
     -> port: STRING(80)
     -> alternate_port: STRING(0)
     -> command: STRING(nginx -g daemon off;)
     -> container_id: STRING(2567fdbc764706d29120b01efcc3a310d87e9e121ec9debbc977d66f5497cdda)
     -> transport: STRING(TCP)
     -> name: STRING(ecstatic_davinci)
     -> image: STRING(nginx)
     -> host: STRING(172.17.0.2)
     -> endpoint: STRING(172.17.0.2:80)
Trace ID:
Span ID:
Flags: 0
LogRecord #2
ObservedTimestamp: 2022-07-27 19:00:32.305784658 +0000 UTC
Timestamp: 2022-07-27 19:00:32.305784612 +0000 UTC
Severity: info
Body: {"alternate_port":0,"command":"redis-server","container_id":"5f7f5f007f798c59a60c765e566db5d22bff59c614268db8d1b9abbc3ee70bf7","endpoint":"172.17.0.3:6379","host":"172.17.0.3","id":"5f7f5f007f798c59a60c765e566db5d22bff59c614268db8d1b9abbc3ee70bf7:6379","image":"redis","labels":{},"name":"beautiful_clarke","port":6379,"tag":"latest","transport":"TCP","type":"container"}
Attributes:
     -> type: STRING(endpoint)
     -> command: STRING(redis-server)
     -> transport: STRING(TCP)
     -> endpoint: STRING(172.17.0.3:6379)
     -> port: STRING(6379)
     -> image: STRING(redis)
     -> tag: STRING(latest)
     -> alternate_port: STRING(0)
     -> container_id: STRING(5f7f5f007f798c59a60c765e566db5d22bff59c614268db8d1b9abbc3ee70bf7)
     -> host: STRING(172.17.0.3)
     -> labels: STRING(map[])
     -> id: STRING(5f7f5f007f798c59a60c765e566db5d22bff59c614268db8d1b9abbc3ee70bf7:6379)
     -> name: STRING(beautiful_clarke)
Trace ID:
Span ID:
Flags: 0
```

## Example Usage

The following Collector configuration will create a Discovery receiver instance that receives
endpoints from a [Docker Observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/observer/dockerobserver/README.md)
that reports log records denoting the status of a [Smart Agent receiver](https://github.com/signalfx/splunk-otel-collector/blob/main/pkg/receiver/smartagentreceiver/README.md)
configured to use the `collectd/redis` monitor. The `status` mapping comprises entries that signal the Smart Agent
receiver has been instantiated with a `successful`, `partial`, or `failed` status, based on reported `metrics` or
recorded application log `statements`.

The Redis receiver warrants the following log record with `discovery.status` from the Discovery receiver:

* `successful` if it emits any metrics matching the `regexp: .*` pattern, denoting that metric gathering and the
Receiver are functional.
* `partial` if it internally logs a statement matching the `regexp: (WRONGPASS|NOAUTH|ERR AUTH)` pattern, suggesting
there is a Redis server but it's receiving incorrect credentials.
* `failed` if it internally logs a statement matching the `regexp: ConnectionRefusedError` pattern, suggesting that no
Redis server is available at the endpoint.

```yaml
extensions:
  docker_observer:
receivers:
   discovery:
     watch_observers: [docker_observer]
     receivers:
       smartagent/redis:
         // Determine the functionality of the smartagent/redis receiver for any detected container
         rule: type == "container"
         config:
           type: collectd/redis
           // the metric or log statement
         status:
           metrics:
             successful:
               - regexp: '.*'
                 // Only emit a single log record for this status entry instead of one for each matching received metric (`false`, the default)
                 first_only: true
                 log_record:
                   severity_text: info
                   body: Successfully able to connect to Redis container.
           statements:
             partial:
               - regexp: (WRONGPASS|NOAUTH|ERR AUTH)
                 first_only: true
                 log_record:
                   severity_text: warn
                   body: Container appears to be accepting redis connections but the default auth setting is incorrect.
             failed:
               - regexp: ConnectionRefusedError
                 first_only: true
                 log_record:
                   severity_text: debug
                   body: Container appears to not be accepting redis connections.
exporters:
  logging:
    verbosity: detailed
    sampling_initial: 1
    sampling_thereafter: 1
service:
  extensions:
    - docker_observer
  pipelines:
    logs:
      receivers:
        - discovery
      exporters:
        - logging
```

Given this configuration, if the Discovery receiver's Docker observer instance reports an active Redis container, and
the Smart Agent receiver's associated `collectd/redis` monitor is able to generate metrics for the container, the
receiver will emit something similar to the following log records:

```
2022-07-27T16:35:03.575Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "logging", "#logs": 1}
2022-07-27T16:35:03.575Z	info	ResourceLog #0
Resource SchemaURL:
Resource labels:
     -> container.name: STRING(lucid_goldberg)
     -> container.image.name: STRING(redis)
     -> discovery.receiver.rule: STRING(type == "container" && image == "redis")
     -> discovery.endpoint.id: STRING(2f0d88c1d93b2aafce4a725de370005f9e7e961144551a3df37a88b27ebed48f:6379)
     -> discovery.receiver.name: STRING(smartagent/redis)
     -> event.type: STRING(metric.match)
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
LogRecord #0
ObservedTimestamp: 2022-07-27 16:35:03.575234252 +0000 UTC
Timestamp: 2022-07-27 16:35:01.522543616 +0000 UTC
Severity: info
Body: Successfully able to connect to Redis container.
Attributes:
     -> metric.name: STRING(counter.expired_keys)
     -> discovery.status: STRING(successful)
Trace ID:
Span ID:
Flags: 0
```

Instead, if the Docker observer reports an active Redis container but the `collectd/redis` authentication information is
incorrect, the Discovery receiver will emit something similar to the following log record:

```
2022-07-27T17:11:27.271Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "logging", "#logs": 1}
2022-07-27T17:11:27.271Z	info	ResourceLog #0
Resource SchemaURL:
Resource labels:
     -> event.type: STRING(statement.match)
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
LogRecord #0
ObservedTimestamp: 2022-07-27 17:11:27.27149884 +0000 UTC
Timestamp: 2022-07-27 17:11:27.271444488 +0000 UTC
Severity: warn
Body: Container appears to be accepting redis connections but the default auth is incorrect.
Attributes:
     -> runnerPID: STRING(119386)
     -> monitorID: STRING(smartagentredisreceiver_creatordiscoveryendpoint1721702637958d3f3335daf83f7)
     -> monitorType: STRING(collectd/redis)
     -> caller: STRING(signalfx/handler.go:189)
     -> name: STRING(smartagent/redis/receiver_creator/discovery{endpoint="172.17.0.2:6379"}(58d3f3335daf83f7))
     -> createdTime: STRING(1.658941887269115e+09)
     -> lineno: STRING(201)
     -> logger: STRING(root)
     -> sourcePath: STRING(/usr/lib/splunk-otel-collector/agent-bundle/collectd-python/redis/redis_info.py)
     -> level: STRING(error)
     -> receiver.name: STRING(smartagent/redis)
     -> discovery.status: STRING(partial)
Trace ID:
Span ID:
Flags: 0
```

If the Docker observer reports an unrelated container that isn't running Redis, the following log record would be emitted:

```
2022-07-27T17:16:57.718Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "logging", "#logs": 1}
2022-07-27T17:16:57.719Z	info	ResourceLog #0
Resource SchemaURL:
Resource labels:
     -> event.type: STRING(statement.match)
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
LogRecord #0
ObservedTimestamp: 2022-07-27 17:16:57.718720008 +0000 UTC
Timestamp: 2022-07-27 17:16:57.718670678 +0000 UTC
Severity: debug
Body: Container appears to not be accepting redis connections.
Attributes:
     -> name: STRING(smartagent/redis/receiver_creator/discovery{endpoint="172.17.0.2:80"}(54141aa1da2d2ad0))
     -> monitorType: STRING(collectd/redis)
     -> logger: STRING(root)
     -> createdTime: STRING(1.6589422177182763e+09)
     -> level: STRING(error)
     -> monitorID: STRING(smartagentredisreceiver_creatordiscoveryendpoint17217028054141aa1da2d2ad0)
     -> runnerPID: STRING(119590)
     -> sourcePath: STRING(/usr/lib/splunk-otel-collector/agent-bundle/collectd-python/redis/redis_info.py)
     -> lineno: STRING(198)
     -> caller: STRING(signalfx/handler.go:189)
     -> receiver.name: STRING(smartagent/redis)
     -> discovery.status: STRING(failed)
Trace ID:
Span ID:
Flags: 0
```

## Config

### Main

| Name                         | Type                      | Default    | Docs                                                                                                                                                                                                 |
|------------------------------|---------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `watch_observers` (required) | []string                  | <no value> | The array of Observer extensions to receive Endpoint events from                                                                                                                                     |
| `log_endpoints`              | bool                      | false      | Whether to emit log records for Observer Endpoint events                                                                                                                                             |
| `embed_receiver_config`      | bool                      | false      | Whether to embed a base64-encoded, minimal Receiver Creator config for the generated receiver as a reported metrics `discovery.receiver.rule` resource attribute value for status log record matches |
| `receivers`                  | map[string]ReceiverConfig | <no value> | The mapping of receiver names to their Receiver sub-config                                                                                                                                           |

### ReceiverConfig

| Name                  | Type              | Default    | Docs                                                                                                                                            |
|-----------------------|-------------------|------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| `rule` (required)     | string            | <no value> | The Receiver Creator compatible discover rule                                                                                                   |
| `config`              | map[string]any    | <no value> | The receiver instance configuration, including any Receiver Creator endpoint env value expr program value expansion                             |
| `resource_attributes` | map[string]string | <no value> | A mapping of string resource attributes and their (expr program compatible) values to include in reported metrics for status log record matches |
| `status`              | map[string]Match  | <no value> | A mapping of `metrics` and/or `statements` to Match items for status evaluation                                                                 |

### Match

**One of `regexp`, `strict`, or `expr` is required.**

| Name         | Type      | Default    | Docs                                                                                                                |
|--------------|-----------|------------|---------------------------------------------------------------------------------------------------------------------|
| `strict`     | string    | <no value> | The string literal to compare equivalence against reported received metric names or component log statement message |
| `regexp`     | string    | <no value> | The regexp pattern to evaluate reported received metric names or component log statements                           |
| `expr`       | string    | <no value> | The expr program run with the reported received metric names or component log statements                            |
| `first_only` | bool      | false      | Whether to emit only one log record for the first matching metric or log statement, ignoring all subsequent matches |
| `record`     | LogRecord | <no value> | The emitted log record content                                                                                      |

#### `strict`

For metrics, the metric name must match exactly.
For logged statements, the message (`zapLogger.Info("<this statement message>")`) must match exactly.

#### `regexp`

For metrics, the regexp is evaluated against the metric name.
For logged statements, the regexp is evaluated against the message and fields (`zapLogger.Info("<logged statement message>", zap.Any("field_name", "field_value"))`) rendered as a yaml mapping. The fields for `caller`, `name`, and `stacktrace` are currently withheld from the mapping.

#### `expr`

See [https://expr.medv.io/](https://expr.medv.io/) for env and language documentation.

For metrics, the expr env consists of `{ "name": "<metric name>" }`.
For logs, the expr env consists of `{ "message": "<logged statement message>", "<field_name>": "<field_value>" }`. The fields `caller`, `name`, and `stacktrace` are currently withheld from the env.

Since some fields may not be valid expr identifiers (containing non word characters), the env contains a self-referential `ExprEnv` object:

```go
logger.Warn("some message", zap.String("some.field.with.periods", "some.value"))
```

In this case `some.field.with.periods` can be referenced via:

```yaml
expr: 'ExprEnv["some.field.with.periods"] contains "value"'
```

### LogRecord

| Name             | Type              | Default                                                 | Docs                                                                        |
|------------------|-------------------|---------------------------------------------------------|-----------------------------------------------------------------------------|
| `severity_text`  | string            | Emitted log statement severity level, if any, or "info" | The emitted log record's severity text                                      |
| `body`           | string            | Emitted log statement message                           | The emitted log record's body                                               |
| `attributes`     | map[string]string | Emitted log statements fields                           | The emitted log record's attributes                                         |
| `append_pattern` | bool              | false                                                   | Whether to append the evaluated statement to the configured log record body |

## Status log record content

In addition to the effects of the configured values, each emitted log record will include:

* `event.type` resource attribute with either `metric.match` or `statement.match` based on context.
* `discovery.status` log record attribute with `successful`, `partial`, or `failed` status depending on match.

