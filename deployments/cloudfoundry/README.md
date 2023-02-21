# Splunk OpenTelemetry Collector Pivotal Cloud Foundry (PCF) Integrations

Supported Tanzu Application Service (TAS) versions: v2

Unsupported TAS versions: v3

### Cloud Foundry Buildpack

This integration can be used to install and run the Collector as a sidecar to your app.
In this configuration, the Collector will run in the same container as the app.

### Bosh Release

This is a Bosh release of the Splunk Collector. This deploys the Collector to the PCF
environment as a standalone deployment.

### Tanzu Tile

Packaged release of the Splunk Collector that can be integrated into the Ops Manager.