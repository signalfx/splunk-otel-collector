# Zookeeper Config Source

| Status | |
| --- | --- |
| Stability | Alpha |
| Distributions | [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector) |

Use the [Zookeeper](https://zookeeper.apache.org/) config source to retrieve data from
Zookeeper and inject it into your collector configuration.

## Configuration

Under `config_sources:`, use `zookeeper:` or `zookeeper/<name>:` to create a
Zookeeper config source. All options are optional.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `endpoints` | No | `[localhost:2181]` | Zookeeper server addresses. The list must not be empty. |
| `timeout` | No | `10s` | Session timeout after losing a connection; the client can reconnect during this interval without losing its session. |

### Example configuration

Create named instances to use different clusters:

```yaml
config_sources:
  zookeeper:
    endpoints: ["localhost:2181"]
  zookeeper/another_cluster:
    endpoints: ["zk.example.com:2181"]
    timeout: 15s

components:
  component_using_zookeeper:
    token: ${zookeeper:/data/token}
  component_using_zookeeper_another_cluster:
    token: ${zookeeper/another_cluster:/data/token}
```

Here, `components` and the component names stand for valid component sections
and settings.
