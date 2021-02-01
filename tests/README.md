# Splunk OpenTelemetry Collector Integration Tests and Utilities

To assist in vetting and validating the upstream and Splunk Collector distributions of the Collector, this library
provides a set of integration tests and associated utilities.  The general testing pattern this project is geared toward
is:

1. Building the Collector (`make otelcol` or `make all`)
1. Defining your expected [resource metric content](./testutils/README.md#resource-metrics) as a yaml file
([see example](./testutils/testdata/resourceMetrics.yaml))
1. Spin up your target resources as docker containers (TODO) 
1. Stand up an in-memory OTLP Receiver and metric sink capable of detecting if/when desired data are received (TODO).
1. Spin up your Collector as a subprocess configured to report to the this OTLP receiver (TODO)
  
...but if you are interested in something else enhancements and contributions are a great way to ensure this library
is more useful overall.

At this time only limited metric content is supported.  If you need additional metric functionality or trace/log
helpers, please don't hesitate to contribute!
