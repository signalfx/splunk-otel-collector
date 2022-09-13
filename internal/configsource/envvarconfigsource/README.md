# Environment Variable Config Source (Alpha)

Use the environmental variable config source instead of direct references to
environment variables in the config to inject YAML fragments or to have default
values in case the selected environment variable is undefined. For simple environment
variable expansion without support for YAML fragments or defaults see
[Collector Configuration Environment Variables](https://opentelemetry.io/docs/collector/configuration/#configuration-environment-variables) 

## Configuration

Under the `config_sources:` use `env:` or `env/<name>:` to create an
environment variable config source. The following parameters are available to
customize environment variable config sources:

```yaml
config_sources:
  env:
    # defaults is used to create a set of fallbacks in case the original env var is
    # undefined in the environment.
    defaults:
      MY_ENV_VAR: my env var value
```

By default, the config source will cause an error if it tries to inject an environment variable
that is not defined or not specified on the `defaults` section. That behavior can be controlled
via the `optional` parameters when invoking the config source, example:

```yaml
config_sources:
  env:
    defaults:
      BACKED_BY_DEFAULTS_ENV_VAR: some_value

components:
  component_0:
    # Not an error if ENV_VAR_NAME is undefined since 'optional' is set to true,
    # the resulting value is "/data/token".
    not_required_field: ${env:ENV_VAR_NAME?optional=true}/data/token 

  component_1:
    # It will be an error if ENV_VAR_NAME is undefined, the config will fail to load.
    required_field: ${env:ENV_VAR_NAME}/data/token 

  component_2:
    # Not an error if BACKED_BY_DEFAULTS_ENV_VAR is undefined, because the 'defaults'
    # of the config source.
    required_field: ${env:BACKED_BY_DEFAULTS_ENV_VAR}/data/token 
```

## Injecting YAML Fragments

The typical case to use the environment variable config source is when one wants
to inject YAML fragments. The example below shows how this can be done on Linux and
Windows when running the collector from your current session. For guidance on setting
service environment variables for your installation please see the related
[Linux](../../../docs/getting-started/linux-manual.md#collector-debianrpm-post-install-configuration)
and [Windows](../../../docs/getting-started/windows-installer.md#collector-configuration) installer documentation.

1. Use the `env` config source environment variables in your configuration:
```yaml
config_sources:
  env:
    defaults:
      JAEGER_PROTOCOLS: "{ protocols: { grpc: , } }"
      OTLP_PROTOCOLS: "{ grpc: , }"

receivers:
  jaeger:
    ${env:JAEGER_PROTOCOLS}
  otlp:
    protocols:
      ${env:OTLP_PROTOCOLS}
...
```

2. Export the environment variables for your session before running the collector:
- Linux:
```terminal
export OTLP_PROTOCOLS="{ grpc: , http: , }"
export JAEGER_PROTOCOLS="{ protocols: { grpc: , thrift_binary: , thrift_compact: , thrift_http: , } }"
otelcol --config <your-configuration.yaml>
```

- Windows:
```terminal
set OTLP_PROTOCOLS={ grpc: , http: , }
set JAEGER_PROTOCOLS={ protocols: { grpc: , thrift_binary: , thrift_compact: , thrift_http: , } }
otelcol.exe --config <your-configuration.yaml>
```
