# One histogram OTLP export

This example showcases how the collector can ingest data from a static Prometheus endpoint and send it to a debug endpoint.

The static content exported is a typical histogram metric data point from Istio.

You can enable an additional export to OTLP HTTP with the otlp_http exporter. You will need to configure the exporter with a valid endpoint and ingest token.

The example runs as a Docker Compose deployment. 

To deploy the example, check out this git repository, open a terminal and in this directory type:
```bash
$> docker-compose up
```