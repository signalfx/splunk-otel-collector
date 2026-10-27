# Include Config Source

| Status | |
| --- | --- |
| Stability | Beta |
| Distributions | [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector) |

Use the include config source to inject [golang templates](https://pkg.go.dev/text/template)
or plain files into the configuration. The config source can be used to insert scalar data
or complete YAML sections.

## Configuration

Under `config_sources:`, use `include:` or `include/<name>:` to create an
include config source. Its only option is optional.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `watch_files` | No | `false` | Watch referenced files and reload the configuration when they change. Reloading temporarily interrupts data flow while pipelines restart. |

### Example configuration

```yaml
config_sources:
  include/default:
  include/watch_for_updates:
    watch_files: true

components:
  component_0:
    component_field: ${include/default:/etc/configs/component_field}
  component_1:
    component_field: ${include/watch_for_updates:/etc/configs/my_config}
```

Here, `components`, `component_0`, `component_1`, and `component_field` stand for
valid component sections and settings. The second reference reloads the
configuration when `/etc/configs/my_config` changes; the first does not.

The config source can be used to insert complete sections of the configuration:

```yaml
config_sources:
  include:

receivers: ${include:/etc/configs/receivers.yaml}
exporters: ${include:/etc/configs/exporters.yaml}
service:
  pipelines: ${include:/etc/configs/pipelines.yaml}
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
  component_0: ${include:./templates/component_template?my_glob_pattern=/var/**/*.log&my_format=json}
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
