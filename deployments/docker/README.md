# Docker

Deploy the latest Docker image:

```bash
$ docker run --rm -e SPLUNK_ACCESS_TOKEN=12345 -e SPLUNK_MEMORY_TOTAL_MIB=1024 \
    -e SPLUNK_REALM=us0 -p 13133:13133 -p 14250:14250 -p 14268:14268 -p 55678-55680:55678-55680 \
    -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 9411:9411 -p 9943:9943 \
    --name otelcol quay.io/signalfx/splunk-otel-collector:latest
```

Replace the values for `SPLUNK_ACCESS_TOKEN`, `SPLUNK_REALM`, and `SPLUNK_MEMORY_TOTAL_MIB` in the command
appropriately for your environment.

A corresponding [docker-compose.yml](./docker-compose.yml) for this command is also available. Change to this directory
and run:

```bash
$ docker-compose up
```

Replace the values in [.env](./.env) appropriately for your environment.

See ["Getting Started"](../../docs/getting-started/linux-standalone.md) for details about these environment variables and other
configuration options.

If desired, replace `latest` with a specific version number for the Docker image (e.g. `0.1.0`).
