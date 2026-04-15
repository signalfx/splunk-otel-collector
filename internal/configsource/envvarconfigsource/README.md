# Environment Variable Config Source (Alpha)

*NOTICE* This environment variable config source is deprecated in favor of the [upstream configuration source for environment variables](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider/envprovider).

Please use the `env:ENV_VAR:-default` notation to set the default value of the environment variable.

Use the environmental variable config source instead of direct references to
environment variables in the config to inject YAML fragments or to have default
values in case the selected environment variable is undefined. For simple environment
variable expansion without support for YAML fragments or defaults see
[Collector Configuration Environment Variables](https://opentelemetry.io/docs/collector/configuration/#configuration-environment-variables) 

## Configuration

By default, the config source will cause an error if it tries to inject an environment variable
that is not defined or not specified on the `defaults` section. That behavior can be controlled
via the `optional` parameters when invoking the config source, example:

```yaml

components:
  component_0:
    # Not an error if ENV_VAR_NAME is undefined since 'optional' is set to true,
    # the resulting value is "/data/token".
    not_required_field: ${env:ENV_VAR_NAME?optional=true}/data/token 

  component_1:
    # It will be an error if ENV_VAR_NAME is undefined, the config will fail to load.
    required_field: ${env:ENV_VAR_NAME}/data/token 

  component_2:
    # Not an error if BACKED_BY_DEFAULTS_ENV_VAR is undefined, because the default value
    # is offered inline after the :- separator
    required_field: ${env:BACKED_BY_DEFAULTS_ENV_VAR:-some_value}/data/token 
```
