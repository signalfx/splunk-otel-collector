# Light Prometheus Receiver

| Status                   |              |
| ------------------------ |--------------|
| Stability                | [deprecated] |
| Supported pipeline types | metrics      |
| Distributions            | splunk       |

[deprecated]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#deprecated

:warning: Please use the [Prometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/prometheusreceiver) or [Simple Prometheus](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver) receiver instead.

## Overview

Light Prometheus Receiver is a component that can scrape Prometheus metrics from a Prometheus exporter endpoint and
convert them to OTLP metrics. It is intended to be used as a replacement for the [Simple Prometheus 
Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/simpleprometheusreceiver)
as it is more efficient and has a smaller memory footprint.

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
  - `server.address`:
    - `enabled`: (default: false)
  - `server.port`:
    - `enabled`: (default: false)
  - `url.scheme`:
    - `enabled`: (default: false)
- [HTTP Client Configuration options](https://github.com/open-telemetry/opentelemetry-collector/tree/main/config/confighttp#client-configuration)

The full list of configuration options exposed for this receiver are documented [here](./config.go) with an example [here](./testdata/config.yaml).