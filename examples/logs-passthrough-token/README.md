# Splunk HEC Example

This example shows how to use the collector as a passthrough for a HEC token.

The collector receives a request from a local curl command with an authorization token.

## Example Deployment
Checkout this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```

Once running, the application will log the final log record with examples such as:
```
otelcollector  | xxxx-xx-xxTxx:xx:xx.xxx       info    ResourceLog #0
otelcollector  | Resource SchemaURL: 
otelcollector  | Resource attributes:
otelcollector  |      -> host.name: Str(unknown)
otelcollector  |      -> com.splunk.source: Str(foo-0000-0000-0000-0000000000128)
otelcollector  |      -> com.splunk.index: Str(logs)
otelcollector  |      -> com.splunk.hec.access_token: Str(foo-0000-0000-0000-0000000000128)
otelcollector  | ScopeLogs #0
otelcollector  | ScopeLogs SchemaURL: 
otelcollector  | InstrumentationScope  
otelcollector  | LogRecord #0
otelcollector  | ObservedTimestamp: 1970-01-01 00:00:00 +0000 UTC
otelcollector  | Timestamp: 1970-01-01 00:00:00 +0000 UTC
otelcollector  | SeverityText: 
otelcollector  | SeverityNumber: Unspecified(0)
otelcollector  | Body: Str(event)
otelcollector  | Trace ID: 
otelcollector  | Span ID: 
otelcollector  | Flags: 0
otelcollector  |        {"kind": "exporter", "data_type": "logs", "name": "debug"}
```

You can also perform this on your desktop with:
```
curl -k http://localhost:18088/services/collector -d '{"event": "event"}' -H "Authorization: Splunk 00000000-0000-0000-0000-0000000000123"
curl -k http://localhost:18088/services/collector/raw -d 'my raw event' -H "Authorization: Splunk 00000000-0000-0000-0000-0000000000123"
```
