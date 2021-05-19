# Include Config Source (Alpha)

Use the include config source to inject [golang templates](https://pkg.go.dev/text/template)
or plain files into the configuration.

## Configuration

Under the `config_sources:` use `include:` or `include/<name>:` to create a
template config source.
The following parameters are available to customize Vault config sources:

```yaml
config_sources:
  include:
  include/my_name_00:
    # delete_files can be used to make the "include" config source to delete the
    # files referenced by it. This is typically used to remove secrets from the
    # file system. The file is deleted as soon as its value is read by the config
    # source. The default value is false. It is an invalid configuration to set it
    # to true together with the watch_files parameter (see below).
    delete_files: true
  include/my_name_01:
    # watch_files can be used to make the "include" config source monitor for
    # updates on the used files. Setting it to true will trigger a configuration
    # reload if any of the files used by the config source are updated.
    # Configuration reload causes temporary interruption of the data flow during
    # the time taken to shut down the current pipeline configuration and start the
    # new one. The default value is false. It is an invalid configuration to set it
    # to true together with the delete_files parameter (see above).
    watch_files: true
```

Example on how to use the `delete_files` and `watch_files`:
```yaml
config_sources:
  include/default:
  include/secret:
    delete_files: true
  include/watch_for_updates:
    watch_files: true

components:
  component_0:
    # Default usage: configuration won't be reloaded if the file
    # '/etc/configs/component_field' is changed.
    component_field: ${include/default:/etc/configs/component_field} 

  component_1:
    # 'include/secret' was created with 'delete_files' set to true the
    # file '/etc/configs/secret' after its value is read. If the deletion
    # fails the collector won't start.
    component_field: ${include/secret:/etc/configs/secret} 

  component_2:
    # 'include/watch_for_updates' was created with 'watch_files' set to true.
    # If the file '/etc/configs/my_config' is changed the collector configuration
    # will be reloaded.
    component_field: ${include/watch_for_updates:/etc/configs/my_config} 
```

If the file being included is a [golang template](https://pkg.go.dev/text/template)
the parameters on the specific reference are used to process the template
For example, assuming that `./templates/component_template` looks like:

```terminal
logs_path: {{ .my_glob_pattern }}
log_format: {{ .my_format }}
```

Given the configuration file:

```yaml
config_sources:
  include:

components:
  # component_0 is built from the ./templates/component_template file
  # according to the template parameters and commands. The example below
  # defines a few parameters to be used by the template.
  component_0: |
    $include: ./templates/component_template
    my_glob_pattern: /var/**/*.log
    my_format: json
```

The effective configuration will be:

```yaml
components:
  component_0:
    logs_path: /var/**/*.log
    log_format: json 
```

See [golang templates](https://pkg.go.dev/text/template)
for a complete description of templating functions and syntax.
