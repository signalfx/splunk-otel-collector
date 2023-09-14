# Telemetrygen trace generation example

This example showcases how the collector can generate traces with [telemetrygen, a metric and traces generator](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen).

The example runs as a Docker Compose deployment. The collector is configured to output all traces to stdout.

To deploy the example, check out this git repository, open a terminal and in this directory type:

```bash
$> docker-compose up
```

The output of the collector will look like the following:
```
otelcollector  | Span #367
otelcollector  |     Trace ID       : c4ab75422bd2171f60c8afe8d7751e45
otelcollector  |     Parent ID      : 
otelcollector  |     ID             : 3c44a16edb446d4e
otelcollector  |     Name           : lets-go
otelcollector  |     Kind           : Internal
otelcollector  |     Start time     : 2023-09-13 19:11:44.703129876 +0000 UTC
otelcollector  |     End time       : 2023-09-13 19:11:44.704767876 +0000 UTC
otelcollector  |     Status code    : Unset
otelcollector  |     Status message : 
otelcollector  | Attributes:
otelcollector  |      -> span.kind: Str(client)
otelcollector  |      -> net.peer.ip: Str(1.2.3.4)
otelcollector  |      -> peer.service: Str(telemetrygen-server)
otelcollector  | Span #368
otelcollector  |     Trace ID       : 4ebd51699f8ee4e4f6146aac935c709b
otelcollector  |     Parent ID      : f5cfe582d01ad47b
otelcollector  |     ID             : c7b587a664e15ae7
otelcollector  |     Name           : okey-dokey
otelcollector  |     Kind           : Internal
otelcollector  |     Start time     : 2023-09-13 19:11:44.704700293 +0000 UTC
```