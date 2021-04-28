# Environment Variable Config Source (Alpha)

Use the environmental variable config source instead of direct references to
environment variables in the config to inject YAML fragments or to have default
values in case the selected environment variable is undefined.

## Configuration

Under the `config_sources:` use `env:` or `env/<name>:` to create an
environment variable config source. The following parameters are available to customize
environment variable config sources:

```yaml
config_sources:
  env:
    # defaults is used to create a set of fallbacks in case the original env var is
    # undefined in the environment.
    defaults:
      MY_ENV_VAR: my env var value
```

It is possible to fail the configuration load if an environment variable is required
and must be defined either on the environment or on the `defaults` of the config source.

```yaml
config_sources:
  env:
    defaults:
      BACKED_BY_DEFAULTS_ENV_VAR: some_value

components:
  component_0:
    # Not an error if ENV_VAR_NAME is undefined, resulting value is "/data/token".
    not_required_field: ${env:ENV_VAR_NAME}/data/token 

  component_1:
    # It will be an error if ENV_VAR_NAME is undefined, the config will fail to load.
    required_field: ${env:ENV_VAR_NAME?required=true}/data/token 

  component_2:
    # Not an error if BACKED_BY_DEFAULTS_ENV_VAR is undefined, because the 'defaults'
    # of the config source.
    required_field: ${env:BACKED_BY_DEFAULTS_ENV_VAR?required=true}/data/token 
```
