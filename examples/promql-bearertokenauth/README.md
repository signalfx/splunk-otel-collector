# PromQL receiver with bearer token authentication

This example runs Prometheus and the Splunk OpenTelemetry Collector in Docker Compose. The collector queries Prometheus' `up` metric with the PromQL receiver and sends the resulting metrics to the debug exporter, where they appear in the collector logs.

Prometheus' built-in web authentication supports basic authentication, not bearer-token validation. An NGINX proxy in front of Prometheus checks the `Authorization: Bearer ...` header and forwards authorized requests to the Prometheus API. The collector's `bearertokenauth` extension supplies that header to the PromQL receiver.

Start the example from this directory:

```sh
docker compose up
```

The default token is `prometheus-demo-token`. To use another token, set `PROMETHEUS_BEARER_TOKEN` in the environment before starting Compose; Compose passes the same value to NGINX and the collector.

The Prometheus and proxy services are only reachable from the Compose network. The example uses HTTP for this local network and is intended for demonstration, not production use.
