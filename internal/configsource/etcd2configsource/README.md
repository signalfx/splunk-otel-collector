# Etcd2 Config Source

| Status | |
| --- | --- |
| Stability | Alpha |
| Distributions | [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector) |

Use the [Etcd2](https://etcd.io/docs/v2.3/) config source to retrieve data from
Etcd2 and inject it into your collector configuration.

## Configuration

Under `config_sources:`, use `etcd2:` or `etcd2/<name>:` to create an Etcd2
config source. All options are optional.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `endpoints` | No | `[http://localhost:2379]` | Etcd2 server URLs to try. The list must not be empty. |
| `auth` | No | Unset | Optional username and password authentication. |
| `auth.username` | No | Empty | Etcd2 username. |
| `auth.password` | No | Empty | Password for `auth.username`. |

### Example configuration

Create named instances to use different clusters or credentials:

```yaml
config_sources:
  etcd2:
    endpoints: ["http://localhost:2379"]
  etcd2/withauth:
    endpoints: ["https://etcd.example.com:2379"]
    auth:
      username: ${env:ETCD2_USERNAME}
      password: ${env:ETCD2_PASSWORD}

components:
  component_using_etcd2:
    token: ${etcd2:/data/token}
  component_using_etcd2_withauth:
    token: ${etcd2/withauth:/data/token}
```
