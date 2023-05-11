# SignalFx Gateway Prometheus remote write receiver

This receiver aims to be an otel-native version of our signalfx [prometheus remote write](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go) [gateway](https://github.com/signalfx/gateway/blob/main/README.md).

## Known limitations
This receiver obsoletes the near-exact behavior of the [SignalFx Prometheus Remote-Writegateway](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go). The behavior of the Prometheus Remote-Write gateway predates the formalization of the Prometheus Remote-Write specification version 1, and differs in the following ways:
- The receiver doesn't [remove suffixes](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/6658646e7705b74f13031c777fcd8dd1cd64c850/receiver/prometheusreceiver/internal/metricfamily.go#L316) as this is done in the otel-contrib `prometheusreceiver`.
- The receiver transforms histograms [into counters](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L98).
- The receiver transforms quantiles (summaries) into gauges.
- If the representation of a float can be expressed as an integer without loss, the receiver sets the representation of a float as an integer.
- If the representation of a sample is NaN, the receiver reports an additional counter with the metric name [`"prometheus.total_NAN_samples"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL190C24-L190C53).
- If the representation of a sample is missing a metric name, the receiver reports an additional counter with the metric name [`"prometheus.total_bad_datapoints"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL191C24-L191C24).
- Any errors in parsing the request report an additional counter,  [`"prometheus.invalid_requests"`](https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#LL189C80-L189C91).
- Metadata from the `prompb.WriteRequest` is ignored.
  The following behavior from sfx gateway is not supported:
- `"request_time.ns"` is no longer reported.  `obsreport` handles similar functionality.
- `"drain_size"` is no longer reported.  `obsreport` handles similar functionality.

## Receiver configuration
This receiver is configured through standard OpenTelemetry mechanisms.  See [`config.go`](./config.go) for details.
* `path` is the path in which the receiver responds to prometheus remote-write requests. The default values is `/metrics`.
* `buffer_size` is the degree to which metric translations can be buffered without blocking further write requests. The default value is `100`.
  This receiver uses `opentelemetry-collector`'s [`confighttp`](https://github.com/open-telemetry/opentelemetry-collector/blob/main/config/confighttp/confighttp.go#L206) options if you want to set up TLS and other features. However, the receiver makes the following changes to upstream default options:
* `endpoint` is the default interface and port to listen on. The default value is `localhost:19291`.
 
If everything is configured properly, logs with sample writes should start appearing in stdout shortly.