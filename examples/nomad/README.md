# Nomad demo Example

This example showcases how the collector can send data to a Splunk Enterprise deployment over the Nomad cluster.

The example will deploy the Splunk OpenTelemetry Collector job as `agent` and `gateway` to collect metrics and traces and export them using the `SignalFx` exporter. This will also deploy `load generators` job to generate traces of `Zipkin` and `Jaeger`.

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$ git clone https://github.com/signalfx/splunk-otel-collector.git
$ cd splunk-otel-collector/examples/nomad
$ nomad run otel-demo.nomad
```
