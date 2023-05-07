# SignalFx Gateway Prometheus remote write receiver

## Limitations and Nuances in translation
This receiver specifically obsoletes the near-exact behavior of the [SignalFx prometheus remote write gateway](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go).
The behavior of the prometheus remote write gateway predates the formalization of the PRW v1 specification, and thus differs in the following ways.

- We do not [remove suffixes](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/6658646e7705b74f13031c777fcd8dd1cd64c850/receiver/prometheusreceiver/internal/metricfamily.go#L316) as is done in the otel-contrib `prometheusreceiver`
- It will transform histograms [into counters](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L98).
- It will transform quantiles (summaries) into gauges.
- If the representation of a float could be expressed as an integer without loss, we will set it as an integer
- If the representation of a sample is NAN, we will report an additional counter with the metric name [`"prometheus.total_NAN_samples"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL190C24-L190C53)
- If the representation of a sample is missing a metric name, we will report an additional counter with the metric name [`"prometheus.total_bad_datapoints"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL191C24-L191C24)
- Any errors in parsing the request will report an additional counter [`"prometheus.invalid_requests"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL189C80-L189C91)
- Metadata from the `prompb.WriteRequest` is **ignored**

The following behavior from sfx gateway is not supported
- `"request_time.ns"` is no longer reported.  `obsreport` handles similar functionality.
- `"drain_size"` is no longer reported.  `obsreport` handles similar functionality.

## Receiver Configuration
This receiver is configured via standard OpenTelemetry mechanisms.  See [`config.go`](./config.go) for specific details.

* `path` is the path in which the receiver should respond to prometheus remote write requests.
  * Defaults to `/metrics`
* `buffer_size` is the degree to which metric translations may be buffered without blocking further write requests.
  * Defaults to `100`

This receiver uses `opentelemetry-collector`'s [`confighttp`](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/confighttp.go#L206) options if you would like to set up tls or similar.  (See linked documentation for the most up-to-date details).
However, we make the following changes to their default options:
* `endpoint` is the default interface + port to listen on
  * Defaults to `localhost:19291`

