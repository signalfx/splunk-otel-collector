# SignalFx Smart Agent Configuration Translation Tool (Experimental)

This package provides a command-line tool, `translatesfx`, that translates a 
SignalFx Smart Agent configuration file into a configuration that can be
used by an OpenTelemetry Collector.

## Caveats

This tool has been used against several example configs and has been shown 
in some cases to produce OpenTelemetry Collector configs that are fully 
functional and comparable to the original Smart Agent config. However, 
this tool is designed to produce only a reasonably accurate approximation of 
the final, production Otel config that would replace a Smart Agent config. It is 
not designed to, and cannot, produce a drop-in replacement Otel config for a
Smart Agent one in all cases.

This tool aims to remove a lot of the drudgery from migrating to Otel 
from Smart Agent, but any config produced by this tool should be carefully 
evaluated and tested before being put into production.

## Usage

The `translatesfx` command requires one argument, a Smart Agent 
configuration file, and accepts a second argument, the working directory 
used by any Smart Agent `#from` directives. The `translatesfx` command uses 
this working directory to resolve any relative paths to files referenced by 
any `#from` directives. This working directory argument may be omitted, in 
which case `translatesfx` expands relative file paths using the current 
working directory.

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
