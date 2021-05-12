# File Config Source (Alpha)

Use the file config source to inject YAML fragments or scalars into the
configuration.

## Configuration

Under the `config_sources:` use `file:` or `file/<name>:` to create a
file config source. The following parameters are available to
customize file config sources:

```yaml
config_sources:
  file:
```

By default, the config source will monitor for updates on the used files
and will trigger a configuration reload when they are updated. Optionally,
the config source can be configured to delete the injected file (typically
to remove secrets from the file system) or to not watch for changes to the
file.

```yaml
config_sources:
  file:

components:
  component_0:
    # Default usage: configuration will be reloaded if the file
    # '/etc/configs/component_field' is changed.
    component_field: ${file:/etc/configs/component_field} 

  component_1:
    # Use the 'delete' parameter to force the removal of files with
    # secrets that shouldn't stay in the OS.
    component_field: ${file:/etc/configs/secret?delete=true} 

  component_2:
    # Use the 'disable_watch' parameter to avoid reloading the configuration
    # if the file is changed.
    component_field: ${file:/etc/configs/secret?disable_watch=true} 
```
