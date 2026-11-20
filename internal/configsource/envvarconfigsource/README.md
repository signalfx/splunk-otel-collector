# Environment Variable Config Source

| Status | |
| --- | --- |
| Stability | Alpha |
| Distributions | [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector) |

*NOTICE* This environment variable config source is deprecated in favor of the [upstream configuration source for environment variables](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider/envprovider).

Use `env:ENV_VAR:-default` to provide a fallback value for an unset variable.

Use the environment variable config source instead of direct references to
environment variables in the config to inject YAML fragments or to have default
values in case the selected environment variable is undefined. For simple environment
variable expansion without support for YAML fragments or defaults see
[Collector Configuration Environment Variables](https://opentelemetry.io/docs/collector/configuration/#configuration-environment-variables) 

## Configuration

Under `config_sources:`, use `env:` or `env/<name>:` to create an environment
variable config source. The source requires no options. The legacy `defaults`
option and `optional` reference parameter are deprecated; use the inline
`:-default` form instead.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `defaults` | No | Empty | Deprecated map of fallback values for unset environment variables. Prefer inline `:-default` values. |

| Reference parameter | Required | Default | Description |
| --- | --- | --- | --- |
| `optional` | No | `false` | Deprecated. When `true`, an unset variable without a fallback does not cause an error. |

Without a fallback, an unset variable causes a configuration error.

### Example configuration

```yaml
config_sources:
  env:

components:
  component_0:
    field_with_fallback: ${env:ENV_VAR_NAME:-default_value}/data/token
  component_1:
    required_field: ${env:ENV_VAR_NAME}/data/token
```

Here, `components`, the component names, and the fields stand for valid
component sections and settings. The first reference uses an inline fallback;
the second causes an error if `ENV_VAR_NAME` is unset. Legacy configurations
can use `${env:ENV_VAR_NAME?optional=true}` to suppress that error.
