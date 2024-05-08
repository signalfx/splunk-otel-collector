# Splunk OpenTelemetry Collector Tanzu Tile

This directory is used to generate a Tanzu tile of the [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector).
The Tanzu tile uses the BOSH release to deploy the collector as a [loggregator firehose nozzle](https://docs.vmware.com/en/Tile-Developer-Guide/3.0/tile-dev-guide/nozzle.html).

## Dependencies

The `release` script requires:

- `jq`
- `wget`
- [Bosh CLI](https://bosh.io/docs/cli-v2-install/)
- [Tile Generator](https://docs.vmware.com/en/Tile-Developer-Guide/3.0/tile-dev-guide/tile-generator.html)

## Releasing

```shell
$ ./make-latest-tile
```
This command will create the BOSH release, and package it as a dependency for the tile. It will then generate the
tile with the same version as the collector. If successful, the tile will be found here: 
`./product/splunk-otel-collector-<VERSION>.pivotal`

## Development and Configuration

Refer to the [Development Guide](./DEVELOPMENT.md) for more information.
