# Splunk OpenTelemetry Collector Integration Tests and Utilities

To assist in vetting and validating the upstream and Splunk Collector distributions of the Collector, this library
provides a set of integration tests and associated utilities.  The general testing pattern this project is geared toward
is:

1. Building the Collector (`make otelcol` or `make all`)
1. Defining [your expected golden file content as a yaml file](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/golden)
1. Spin up your target resources as [docker containers](./testutils/README.md#test-containers).
1. Stand up an in-memory [OTLP metrics receiver and sink](./testutils/README.md#otlp-metrics-receiver-sink) capable of detecting if/when desired data are received.
1. Spin up your Collector [as a subprocess](./testutils/README.md#collector-process) or [as a container](./testutils/README.md#collector-container) configured to report to this OTLP receiver.
  
...but if you are interested in something else enhancements and contributions are a great way to ensure this library
is more useful overall.

**NOTE** At this time, integration tests generally target collector containers (`SPLUNK_OTEL_COLLECTOR_IMAGE` env var),
and test coverage for the subprocess is best effort only, unless the test cases explicitly maintain one.
The collector process targets are generally for test development without requiring frequent rebuilds of a local docker image.
