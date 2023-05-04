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
* `sfx_gateway_compatibility` will transmit otel metrics in a similar shape to how the signalfx prometheus gateway does.  Specifically, it will transform histograms and quantiles [into counters](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L98), suffixing the `le` or `quantile` name to the metric name.
  * Defaults to `false`

This receiver uses `opentelemetry-collector`'s [`confighttp`](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/confighttp.go#L206) options if you would like to set up tls or similar.  (See linked documentation for the most up-to-date details).
However, we make the following changes to their default options:
* `endpoint` is the default interface + port to listen on
  * Defaults to `localhost:19291`

## Remote write client configuration
If you're using the [native remote write configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write), it's advisable that you enable `send=true` under `metadata_config`.
If possible, wait on sending multiple requests until you're reasonably assured that metadata has propagated to the receiver.

## Nuances in translation
- We do not [remove suffixes](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/6658646e7705b74f13031c777fcd8dd1cd64c850/receiver/prometheusreceiver/internal/metricfamily.go#L316) as is done in the otel-contrib `prometheusreceiver`
- Keep in mind promethes timestamps are in unix epoch milliseconds, while otel timestamps are in unix epoch nanoseconds

### Signalfx Compatibility Mode
Turning on the `sfx_gateway_compatibility` configuration option will result in the following changes
- It will transform histograms [into counters](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L98), suffixing the `le` to the metric name.
- It will transform quantiles (summaries) [into counters](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L98), suffixing `quantile` to the metric name.
- If the representation of a float could be expressed as an integer without loss, we will set it as an integer
- If the representation of a sample is NAN, we will report an additional counter with the metric name [`"prometheus.total_NAN_samples"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL190C24-L190C53)
- If the representation of a sample is missing a metric name, we will report an additional counter with the metric name [`"prometheus.total_bad_datapoints"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL191C24-L191C24)
- Any errors in parsing the request will report an additional counter [`"prometheus.invalid_requests"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL189C80-L189C91)
- Metadata is IGNORED
 
The following options from sfx gateway will not be translated
- `"request_time.ns"` is no longer reported.  `obsreport` handles similar functionality.
- `"drain_size"` is no longer reported.  `obsreport` handles similar functionality.
