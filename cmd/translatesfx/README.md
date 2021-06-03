# SignalFx Smart Agent Configuration Translation Tool (Experimental)

This package provides a command-line tool to translate a SignalFx Smart Agent
configuration file into an OpenTelemetry Collector configuration.

## Usage

The `translatesfx` command requires one argument, the signalfx configuration
file, and accepts a second argument, the working directory used by any #from
directives. The `translatesfx` command uses the working directory to resolve any
relative paths to files. If you omit the working directory argument, 
`translatesfx` expands relative files paths using the current working
directory.

```
> translatesfx <sfx-file> [<file expansion working directory>]
```

## Examples

#### CLI Usage

###### Using the current working directory:

```
> translatesfx path/to/sfx/config.yaml
```

###### Using a custom working directory:

```
> translatesfx path/to/sfx/config.yaml path/to/sfx
```
#### Example Input/Output

Given the following input
```yaml
signalFxAccessToken: {"#from": "env:SFX_ACCESS_TOKEN"}
ingestUrl: {"#from": "ingest_url", default: "https://ingest.signalfx.com"}
apiUrl: {"#from": "api_url", default: "https://api.signalfx.com"}
traceEndpointUrl: {"#from": 'trace_endpoint_url', default: "https://ingest.signalfx.com/v2/trace"}

intervalSeconds: 10

logging:
  level: info

monitors:
  - {"#from": "monitors/*.yaml", flatten: true, optional: true}
  - type: memory
```

and the following included files:

###### ingest_url
```
https://ingest.us1.signalfx.com
```

###### api_url
```
https://api.us1.signalfx.com
```

###### trace_endpoint_url
```
https://ingest.signalfx.com/v2/trace
```

###### monitors/cpu.yaml
```yaml
- type: cpu
```

###### monitors/load.yaml
```yaml
- type: load
```

`translatesfx` will output:
```yaml
receivers:
  smartagent/cpu:
    type: cpu
  smartagent/load:
    type: load
  smartagent/memory:
    type: memory
exporters:
  signalfx:
    access_token: ${SFX_ACCESS_TOKEN}
    realm: us1
service:
  pipelines:
    metrics:
      receivers:
      - smartagent/cpu
      - smartagent/load
      - smartagent/memory
      exporters:
      - signalfx
```
