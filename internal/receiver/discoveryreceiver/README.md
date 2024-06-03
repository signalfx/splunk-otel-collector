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
records based on the `status: statements` rules you define. The matching rules SHOULD NOT conflict with each other.
The first matching rule in the list will be used to determine the status of the receiver.

The receiver emits entity events for 
[Endpoints](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/observer/endpoints.go)
discovered by the specified `watch_observers` if they match a rule for any configured receiver.

```
2024-04-17T19:53:24.285Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": 1, "log records": 2}
2024-04-17T19:53:24.286Z	info	ResourceLog #0
Resource SchemaURL:
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
InstrumentationScope attributes:
     -> otel.entity.event_as_log: Bool(true)
LogRecord #0
ObservedTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-04-17 19:53:24.28493287 +0000 UTC
SeverityText:
SeverityNumber: Unspecified(0)
Body: Empty()
Attributes:
     -> otel.entity.id: Map({"discovery.endpoint.id":"k8s_observer/ed171efd-f5ab-4bab-923d-d30f3f221367/(9080)"})
     -> otel.entity.event.type: Str(entity_state)
     -> otel.entity.attributes: Map({"discovery.observer.name":"","discovery.observer.type":"k8s_observer","endpoint":"192.168.33.122:9080","name":"","pod":{"annotations":{"kubernetes.io/psp":"eks.privileged"},"labels":{"appv":"reviews","pod-template-hash":"7bff4f6574","version":"v1"},"name":"reviews-v1-7bff4f6574-fbkw9","namespace":"default","uid":"ed171efd-f5ab-4bab-923d-d30f3f221367"},"port":9080,"transport":"TCP","type":"port"})
Trace ID:
Span ID:
Flags: 0
LogRecord #1
ObservedTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-04-17 19:53:24.28493287 +0000 UTC
SeverityText:
SeverityNumber: Unspecified(0)
Body: Empty()
Attributes:
     -> otel.entity.id: Map({"discovery.endpoint.id":"k8s_observer/ea8ee4f5-31a7-48f0-a3c7-ec41e736ccad/jaeger-grpc(14250)"})
     -> otel.entity.event.type: Str(entity_state)
     -> otel.entity.attributes: Map({"discovery.observer.name":"","discovery.observer.type":"k8s_observer","endpoint":"192.168.57.181:14250","name":"jaeger-grpc","pod":{"annotations":{"checksum/config":"d6cc5d07fe24d77b0d0af827295879943d59e87013f0f1e34fa916b942c51336","kubectl.kubernetes.io/default-container":"otel-collector","kubernetes.io/psp":"eks.privileged"},"labels":{"app":"splunk-otel-collector","component":"otel-collector-agent","controller-revision-hash":"6cb7d5c864","pod-template-generation":"5","release":"my-splunk-otel-collector"},"name":"my-splunk-otel-collector-agent-cdv7s","namespace":"default","uid":"ea8ee4f5-31a7-48f0-a3c7-ec41e736ccad"},"port":14250,"transport":"TCP","type":"port"})
Trace ID:
Span ID:
Flags: 0
```

## Example Usage

The following Collector configuration will create a Discovery receiver instance that receives
endpoints from a [Kubernetes Observer](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/observer/k8sobserver/README.md)
that reports log records denoting the status of a [MySQL receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/mysqlreceiver/README.md).
The `status` mapping comprises entries that signal the receiver has been instantiated with a `successful`, `partial`, 
or `failed` status, based on reported `metrics` or recorded application log `statements`.

The following rules are defined for the `mysql` receiver:

* `successful` if it emits any `mysql.locks` metrics, denoting that metric gathering and the Receiver are functional.
* `partial` if it internally logs a statement matching the `Access denied for user` pattern, suggesting
there is a MySQL server but it's receiving incorrect credentials.
* `failed` if it internally logs a statement matching the `Can't connect to MySQL server on .* [(]111[)]` pattern,
suggesting that no MySQL server is available at the endpoint.

```yaml
extensions:
  k8s_observer:
    auth_type: serviceAccount
    node: ${K8S_NODE_NAME}
receivers:
   discovery:
     watch_observers: [k8s_observer]
     receivers:
       mysql:
         rule: type == "port" and port != 33060 and pod.name matches "(?i)mysql"
         config:
           username: root
           password: root
         status:
           metrics:
             - status: successful
               strict: mysql.locks
               message: Mysql receiver is working!
           statements:
             - status: failed
               regexp: "Can't connect to MySQL server on .* [(]111[)]"
               message:  The container cannot be reached by the Collector. The container is refusing MySQL connections.
             - status: partial
               regexp: 'Access denied for user'
               message: >-
                 Make sure your user credentials are correctly specified using the
                 `SPLUNK_DISCOVERY_RECEIVERS_mysql_CONFIG_username="<username>"` and
                 `SPLUNK_DISCOVERY_RECEIVERS_mysql_CONFIG_password="<password>"` environment variables.
exporters:
  debug:
    verbosity: detailed
service:
  extensions:
    - k8s_observer
  pipelines:
    logs:
      receivers:
        - discovery
      exporters:
        - logging
```

Given this configuration, if the Discovery receiver's Kubernetes observer instance reports an active MySQL container, and
the `mysql` receiver is able to generate metrics for the container, the receiver will emit the following entity event:

```
2024-04-08T06:08:58.204Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": 1, "log records": 1}
2024-04-08T06:08:58.204Z	info	ResourceLog #0
Resource SchemaURL:
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
InstrumentationScope attributes:
     -> otel.entity.event_as_log: Bool(true)
LogRecord #0
ObservedTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-04-08 06:08:58.194666193 +0000 UTC
SeverityText:
SeverityNumber: Unspecified(0)
Body: Empty()
Attributes:
     -> otel.entity.id: Map({"discovery.endpoint.id":"k8s_observer/05c6a212-730c-4295-8dd6-0c460c892034/mysql(3306)"})
     -> otel.entity.event.type: Str(entity_state)
     -> otel.entity.attributes: Map({"discovery.event.type":"metric.match","discovery.message":"Mysql receiver is working!","discovery.observer.id":"k8s_observer","discovery.receiver.rule":"type == \"port\" and port != 33060 and pod.name matches \"(?i)mysql\"","discovery.receiver.type":"mysql","discovery.status":"successful","k8s.namespace.name":"default","k8s.pod.name":"mysql-0","k8s.pod.uid":"05c6a212-730c-4295-8dd6-0c460c892034","metric.name":"mysql.locks","mysql.instance.endpoint":"192.168.161.105:3306"})
Trace ID:
Span ID:
Flags: 0
```

Instead, if the Docker observer reports an active MySQL container but the provided authentication information is
incorrect, the Discovery receiver will emit something similar to the following log record:

```
2024-04-08T06:17:36.991Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": 1, "log records": 1}
2024-04-08T06:17:36.992Z	info	ResourceLog #0
Resource SchemaURL:
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
InstrumentationScope attributes:
     -> otel.entity.event_as_log: Bool(true)
LogRecord #0
ObservedTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-04-08 06:17:36.991618675 +0000 UTC
SeverityText:
SeverityNumber: Unspecified(0)
Body: Empty()
Attributes:
     -> otel.entity.id: Map({"discovery.endpoint.id":"k8s_observer/05c6a212-730c-4295-8dd6-0c460c892034/mysql(3306)"})
     -> otel.entity.event.type: Str(entity_state)
     -> otel.entity.attributes: Map({"caller":"mysqlreceiver@v0.97.0/scraper.go:82","discovery.event.type":"statement.match","discovery.message":"Make sure your user credentials are correctly specified using the `--set splunk.discovery.receivers.mysql.config.username=\"\u003cusername\u003e\"` and `--set splunk.discovery.receivers.mysql.config.password=\"\u003cpassword\u003e\"` command or the `SPLUNK_DISCOVERY_RECEIVERS_mysql_CONFIG_username=\"\u003cusername\u003e\"` and `SPLUNK_DISCOVERY_RECEIVERS_mysql_CONFIG_password=\"\u003cpassword\u003e\"` environment variables. (evaluated \"{\\\"error\\\":\\\"Error 1045 (28000): Access denied for user 'root'@'192.168.174.232' (using password: YES)\\\",\\\"kind\\\":\\\"receiver\\\",\\\"message\\\":\\\"Failed to fetch InnoDB stats\\\"}\")","discovery.observer.id":"k8s_observer","discovery.receiver.name":"","discovery.receiver.rule":"type == \"port\" and port != 33060 and pod.name matches \"(?i)mysql\"","discovery.receiver.type":"mysql","discovery.status":"partial","error":"Error 1045 (28000): Access denied for user 'root'@'192.168.174.232' (using password: YES)","kind":"receiver","name":"mysql//receiver_creator/discovery/logs{endpoint=\"192.168.161.105:3306\"}/k8s_observer/05c6a212-730c-4295-8dd6-0c460c892034/mysql(3306)","stacktrace":"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver.(*mySQLScraper).scrape\n\tgithub.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver@v0.97.0/scraper.go:82\ngo.opentelemetry.io/collector/receiver/scraperhelper.ScrapeFunc.Scrape\n\tgo.opentelemetry.io/collector/receiver@v0.97.0/scraperhelper/scraper.go:20\ngo.opentelemetry.io/collector/receiver/scraperhelper.(*controller).scrapeMetricsAndReport\n\tgo.opentelemetry.io/collector/receiver@v0.97.0/scraperhelper/scrapercontroller.go:194\ngo.opentelemetry.io/collector/receiver/scraperhelper.(*controller).startScraping.func1\n\tgo.opentelemetry.io/collector/receiver@v0.97.0/scraperhelper/scrapercontroller.go:169"})
Trace ID:
Span ID:
Flags: 0
```

If the Kubernetes observer reports an unrelated container that isn't running MySQL, the following entity event would be emitted:

```
2024-04-08T07:06:49.502Z	info	LogsExporter	{"kind": "exporter", "data_type": "logs", "name": "debug", "resource logs": 1, "log records": 1}
2024-04-08T07:06:49.502Z	info	ResourceLog #0
Resource SchemaURL:
ScopeLogs #0
ScopeLogs SchemaURL:
InstrumentationScope
InstrumentationScope attributes:
     -> otel.entity.event_as_log: Bool(true)
LogRecord #0
ObservedTimestamp: 1970-01-01 00:00:00 +0000 UTC
Timestamp: 2024-04-08 07:06:49.502297226 +0000 UTC
SeverityText:
SeverityNumber: Unspecified(0)
Body: Empty()
Attributes:
     -> otel.entity.id: Map({"discovery.endpoint.id":"k8s_observer/05c6a212-730c-4295-8dd6-0c460c892034/mysql(3306)"})
     -> otel.entity.event.type: Str(entity_state)
     -> otel.entity.attributes: Map({"caller":"receivercreator@v0.97.0/observerhandler.go:96","discovery.event.type":"statement.match","discovery.message":"The container cannot be reached by the Collector. The container is refusing MySQL connections. (evaluated \"{\\\"endpoint\\\":\\\"192.168.161.105:3306\\\",\\\"endpoint_id\\\":\\\"k8s_observer/05c6a212-730c-4295-8dd6-0c460c892034/mysql(3306)\\\",\\\"kind\\\":\\\"receiver\\\",\\\"message\\\":\\\"starting receiver\\\"}\")","discovery.observer.id":"k8s_observer","discovery.receiver.name":"","discovery.receiver.rule":"type == \"port\" and port != 33060 and pod.name matches \"(?i)mysql\"","discovery.receiver.type":"mysql","discovery.status":"failed","endpoint":"192.168.161.105:3306","endpoint_id":"k8s_observer/05c6a212-730c-4295-8dd6-0c460c892034/mysql(3306)","kind":"receiver","name":"mysql"})
Trace ID:
Span ID:
Flags: 0
```

## Config

### Main

| Name                         | Type                      | Default    | Docs                                                                                                                                                                                                 |
|------------------------------|---------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `watch_observers` (required) | []string                  | <no value> | The array of Observer extensions to receive Endpoint events from                                                                                                                                     |
| `embed_receiver_config`      | bool                      | false      | Whether to embed a base64-encoded, minimal Receiver Creator config for the generated receiver as a reported metrics `discovery.receiver.rule` resource attribute value for status log record matches |
| `receivers`                  | map[string]ReceiverConfig | <no value> | The mapping of receiver names to their Receiver sub-config                                                                                                                                           |

### ReceiverConfig

| Name                  | Type              | Default    | Docs                                                                                                                                                                                              |
|-----------------------|-------------------|------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `rule` (required)     | string            | <no value> | The Receiver Creator compatible discover rule. Ensure that rules defined in different receivers cannot match the same endpoint. Endpoints matching rules from multiple receivers will be ignored. |
| `config`              | map[string]any    | <no value> | The receiver instance configuration, including any Receiver Creator endpoint env value expr program value expansion                                                                               |
| `resource_attributes` | map[string]string | <no value> | A mapping of string resource attributes and their (expr program compatible) values to include in reported metrics for status log record matches                                                   |
| `status`              | map[string]Match  | <no value> | A mapping of `metrics` and/or `statements` to Match items for status evaluation                                                                                                                   |

### Match

**One of `regexp`, `strict`, or `expr` is required.**

| Name         | Type      | Default    | Docs                                                                                                                |
|--------------|-----------|------------|---------------------------------------------------------------------------------------------------------------------|
| `strict`     | string    | <no value> | The string literal to compare equivalence against reported received metric names or component log statement message |
| `regexp`     | string    | <no value> | The regexp pattern to evaluate reported received metric names or component log statements                           |
| `expr`       | string    | <no value> | The expr program run with the reported received metric names or component log statements                            |
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

| Name             | Type              | Default                       | Docs                                |
|------------------|-------------------|-------------------------------|-------------------------------------|
| `body`           | string            | Emitted log statement message | The emitted log record's body       |
| `attributes`     | map[string]string | Emitted log statements fields | The emitted log record's attributes |

## Status log record content

In addition to the effects of the configured values, each emitted log record will include:

* `event.type` resource attribute with either `metric.match` or `statement.match` based on context.
* `discovery.status` log record attribute with `successful`, `partial`, or `failed` status depending on match.

