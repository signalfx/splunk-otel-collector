# Prometheus remote write example

This example provides a `docker-compose` environment that continually sends fake prometheus remote writes to an OTel receiver replacement for the deprecated SignalFx Gateway for Prometheus Remote Writes. To run the example, make sure you have `docker-compose` installed.

## Configuration
To change the exporters, edit the `otel-collector-config.yaml` configuration file.

If you want to send data to Splunk Observability Cloud, set the following environment variables:
- `SPLUNK_ACCESS_TOKEN`
- `SPLUNK_REALM`
 
You can also remove the `signalfx` array item from the `exporters` configuration map in the `otel-collector-config.yaml` configuration file.
> **Tip:** Experiment and modify the sample client, or even disable it and write your own.
 
## Running this example
After you've verified your environment, run the example via `docker-compose up`

If that doesn't work, try running `cd ../../ && make docker-otelcol && cd - && docker-compose up --build`
