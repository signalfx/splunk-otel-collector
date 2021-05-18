# Template Config Source (Alpha)

Use the template config source to inject [golang templates](https://pkg.go.dev/text/template)
into the configuration.

## Configuration

Under the `config_sources:` use `template:` or `template/<name>:` to create a
template config source.

```yaml
config_sources:
  template:
```

The parameters of a template config source are passed to template to be processed.
For example, assuming that `./templates/component_template` looks like:

```terminal
logs_path: {{ .my_glob_pattern }}
log_format: {{ .my_format }}
```

Given the configuration file:

```yaml
config_sources:
  template:

components:
  # component_0 is built from the ./templates/component_template file
  # according to the template parameters and commands. The example below
  # defines a few parameters to be used by the template.
  component_0: |
    $template: ./templates/component_template
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
