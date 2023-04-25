# Prometheus remote write receiver

This prometheus remote write receiver aims to 
1. Deprecate the Prometheus Gateway from signalfx
2. Support prometheus remote writes as an ingestion mechanism to the open-telemetry collector
3. Support flaky clients to the best possible degree

## Limitations
As of this writing, no official specification exists for remote write endpoints, nor a 1-1 mapping between prometheus remote write metrics and OpenTelemetry metrics.

As such, this receiver implements a best-effort mapping between such metrics.  If you find your use case or access patterns do not jive well with this receiver, please [cut an issue](https://github.com/signalfx/splunk-otel-collector/issues/new) to our repo with the specific data incongruity that you're experiencing, and we will do our best to provide for you within maintainable reason.

## Receiver Configuration
This receiver is configured via standard OpenTelemetry mechanisms.  See [`config.go`](./config.go) for specific details.

* `path` is the path in which the receiver should respond to prometheus remote write requests.
  * Defaults to `/metrics`
* `buffer_size` is the degree to which metric translations may be buffered without blocking further write requests.
  * Defaults to `100`
* `cache_size` is the number of most recent metadata requests which should be stored.  Turn this to zero if you wish to disable caching between requests, but do ensure your metrics write requests are independently and consistently parseable without any metadata if so.
  * Defaults to `10000`

This receiver uses `opentelemetry-collector`'s [`confighttp`](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/confighttp.go#L206) options if you would like to set up tls or similar.  (See linked documentation for the most up-to-date details).
However, we make the following changes to their default options:
* `endpoint` is the default interface + port to listen on
  * Defaults to `localhost:19291`

## Remote write client configuration
If you're using the [native remote write configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write), it's advisable that you enable `send=true` under `metadata_config`.
If possible, wait on sending multiple requests until you're reasonably assured that metadata has propagated to the receiver.

## Nuances in translation
We do not [remove suffixes](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/6658646e7705b74f13031c777fcd8dd1cd64c850/receiver/prometheusreceiver/internal/metricfamily.go#L316) as is done in the otel-contrib `prometheusreceiver`
