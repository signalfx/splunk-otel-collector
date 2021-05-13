# YAML Config Source (Alpha)

Use the YAML config source to inject [golang templates](https://pkg.go.dev/text/template)
into the configuration.

## Configuration

Under the `config_sources:` use `yaml:` or `yaml/<name>:` to create a
YAML config source.

```yaml
config_sources:
  yaml:
```

The parameters of an YAML config source are passed to template to be processed.
For example, assumint that `./templates/component_template` looks like:

```terminal
logs_path: {{ .glob_pattern }}
log_format: {{ .format }}
```

Given the configuration file:

```yaml
config_sources:
  yaml:

components:
  # component_0 is built from the ./templates/component_template file
  # according to the template parameters and commands. The example below
  # defines a few parameters to be used by the template.
  component_0: |
    $yaml: ./templates/component_template
    glob_pattern: /var/**/*.log
    format: json
```

The effective configuration will be:

```yaml
config_sources:
  yaml:

components:
  component_0:
    logs_path: /var/**/*.log
    log_format: json 
```

See [golang templates](https://pkg.go.dev/text/template)
for a complete description of templating functions and syntax.
