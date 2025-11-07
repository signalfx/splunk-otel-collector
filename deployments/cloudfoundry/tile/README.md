# Splunk OpenTelemetry Collector Tanzu Tile

This directory is used to generate a Tanzu tile of the [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector).
The Tanzu tile uses the BOSH release to deploy the collector as a [loggregator firehose nozzle](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/nozzle.html).

## CI Build and Artifacts

If you just want to publish a new version of the tile, you can use the artifacts of the
[Tanzu Tile](../../../.github/workflows/tanzu-tile.yml) GitHub Action workflow.

The artifacts of the workflow include:

- The generated tile file: `splunk-otel-collector-<VERSION>.pivotal` at
  `deployments/cloudfoundry/tile/product/`
- The OSDF file, `OSDF V<VERSION>.txt` at the artifact root.
- The `blobs.yml` file at `deployments/cloudfoundry/bosh/config` which contains the
  SHA256 checksum of the collector `pivotal` file. This SHA256 will be used as the content
  of a file named `splunk-otel-collector-<VERSION>.pivotal.sha.txt` when creating
  a new release at the [ISV Tanzu website](https://auth.isv.ci/products/splunk-opentelemetry-collector-for-vmware-tanzu).

## Local Build

__Check [Tanzu Tile](../../../.github/workflows/tanzu-tile.yml) GitHub Action workflow
for the latest versions of the required dependencies.__

```shell
./make-latest-tile
```

This command will create the BOSH release, and package it as a dependency for the tile. It will then generate the
tile with the same version as the collector. If successful, the tile will be found here:
`./product/splunk-otel-collector-<VERSION>.pivotal`

## Development and Configuration

Refer to the [Development Guide](./DEVELOPMENT.md) for more information.
