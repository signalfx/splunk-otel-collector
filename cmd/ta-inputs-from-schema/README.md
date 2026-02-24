# ta-inputs-from-schema

A build tool that generates Splunk modular input configuration files (`inputs.conf` and `inputs.conf.spec`) from an XML scheme file.

## Usage

```bash
go run ./cmd/ta-inputs-from-schema \
  -scheme <path-to-scheme.xml> \
  -global-settings <path-to-global-settings.txt> \
  -name <modular-input-name> \
  -assets <path-to-assets-directory>
```

### Parameters

- `-scheme`: Path to the XML scheme file (required)
- `-global-settings`: Path to the global settings file containing Splunk input configuration (required)
- `-name`: Name of the modular input (required)
- `-assets`: Path to the assets directory where files will be generated (required)

### Example

```bash
go run ./cmd/ta-inputs-from-schema \
  -scheme cmd/otelcol/ta_scheme.xml \
  -global-settings cmd/ta-inputs-from-schema/testdata/global_settings.txt \
  -name Splunk_TA_OTel_Collector \
  -assets packaging/ta-v2/assets
```

This will generate:
- `packaging/ta-v2/assets/default/inputs.conf`
- `packaging/ta-v2/assets/README/inputs.conf.spec`

## Input Files

### XML Scheme File

The XML scheme file defines the modular input arguments. Example:

```xml
<scheme>
    <title>Splunk Add-on for OpenTelemetry Collector</title>
    <description>Deploys the Splunk OpenTelemetry Collector as a modular input</description>
    <endpoint>
        <args>
            <arg name="splunk_access_token" defaultValue="">
                <title>Splunk Access Token</title>
                <description>Access token used to send data to Splunk Observability</description>
                <data_type>string</data_type>
                <required_on_edit>true</required_on_edit>
            </arg>
        </args>
    </endpoint>
</scheme>
```

### Global Settings File

A plain text file containing Splunk input global settings. Example:

```
# Global settings
disabled=false
start_by_shell=false
interval = 0
index = _internal
sourcetype = Splunk_TA_OTel_Collector
```

## Output Files

### inputs.conf

Generated at `<assets>/default/inputs.conf`, this file contains the default configuration for the modular input with all arguments set to their default values.

### inputs.conf.spec

Generated at `<assets>/README/inputs.conf.spec`, this file contains the specification describing all available configuration parameters, their requirements, and default values.

## Testing

Run the unit tests:

```bash
go test ./cmd/ta-inputs-from-schema/... -v
```

## Development

The tool is structured into several files:

- `main.go`: Entry point and command-line argument parsing
- `parser.go`: XML scheme parsing
- `generator.go`: Configuration file generation
- `parser_test.go`: Tests for XML parsing
- `generator_test.go`: Tests for file generation
