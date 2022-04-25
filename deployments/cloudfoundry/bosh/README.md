# Splunk OpenTelemetry Collector BOSH Release

This directory contains a BOSH Release of the [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector).
The BOSH release can be used to deploy the collector so that it will act as a [Loggregator Firehose nozzle](https://docs.pivotal.io/tiledev/2-2/nozzle.html).

## Releasing

See the [`release`](./release) script in this directory. This script can be used to
generate a new release with the latest version of Splunk's distribution of the
OpenTelemetry Collector. This script is recommended to be run from the Pivotal
Cloud Foundry (PCF) tile.

## BOSH Release Usage

```shell
$ bosh -d splunk-otel-collector deploy deployment.yaml
```
Example `deployment.yaml` file included [here.](./example/deployment.yaml)

To include a custom collector configuration file for the deployment, refer to
the `custom_config_deployment.yaml` file included
[here.](./example/custom_config_deployment.yaml)

## Dependencies

The `release` script requires:

- [Bosh CLI](https://bosh.io/docs/cli-v2-install/)
- `wget`
- `jq`

## Development and Configuration

Refer to the [Development Guide](./DEVELOPMENT.md) for more information.
