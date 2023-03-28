# Splunk OpenTelemetry Collector Pivotal Cloud Foundry (PCF) Integrations

Supported Tanzu Application Service (TAS) versions: v2

Unsupported TAS versions: v3

### Cloud Foundry Buildpack

This integration can be used to install and run the Collector as a sidecar to your app.
In this configuration, the Collector will run in the same container as the app.

### Bosh Release

This is a Bosh release of the Collector. This deploys the Collector to the PCF
environment as a standalone deployment.

### Tanzu Tile

This is a Tanzu Tile of the Collector, which is a packaged release of the collector
that can be integrated into the Ops Manager. The Tanzu Tile enables users to download, install,
run, configure, and update the collector all from the Ops Manager.

[Tanzu Tile UI](./tile/resources/tanzu_tile_in_ops_mgr.png)

[Tanzu Tile Configuration UI](./tile/resources/tanzu_tile_config_options.png)