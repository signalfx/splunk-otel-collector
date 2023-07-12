# Light Prometheus Receiver

| Status                   |               |
| ------------------------ |---------------|
| Stability                | [development] |
| Supported pipeline types | metrics       |
| Distributions            | splunk        |

[development]: https://github.com/open-telemetry/opentelemetry-collector#development

## Overview

Light Prometheus Receiver is a component that can scrape Prometheus metrics from a Prometheus exporter endpoint and
convert them to OTLP metrics. It is intended to be used as a replacement for the [Simple Prometheus 
Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver)
as it is more efficient and has a smaller memory footprint.

The receiver is under active development which means that configuration interface can change.

## Configuration

The following settings are required:

- `endpoint` (no default): Address to request Prometheus metrics. This is the same endpoint that 
  Prometheus scrapes to collect metrics. IMPORTANT: This receiver currently does require the metric path to be included
  in the endpoint. For example, if the endpoint is `localhost:1234`, the metrics path must be included, e.g.
  `localhost:1234/metrics`. This likely will be changed in the future.

The following settings can be optionally configured:

- `collection_interval` (default = 30s): The internal at which metrics should be scraped by this receiver.
- `resource_attributes`: Resource attributes to be added to all metrics emitted by this receiver. The following options
  are available to configure resource attributes:
  - `service.name`:
    - `enabled`: (default: true)
  - `service.instance.id`:
    - `enabled`: (default: true)
  - `net.host.name`:
    - `enabled`: (default: false)
  - `net.host.port`:
    - `enabled`: (default: false)
  - `http.scheme`:
    - `enabled`: (default: false)
- [HTTP Client Configuration options](https://github.com/open-telemetry/opentelemetry-collector/tree/main/config/confighttp#client-configuration)

The full list of configuration options exposed for this receiver are documented [here](./config.go) with an example [here](./testdata/config.yaml).