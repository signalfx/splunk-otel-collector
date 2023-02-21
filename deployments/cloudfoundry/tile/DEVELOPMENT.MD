# Development Guide

## VMware Tanzu Tile Documentation

[Tanzu Tile Introduction](https://docs.pivotal.io/tiledev/2-10/tile-basics.html)

[How Tiles Work](https://docs.pivotal.io/tiledev/2-10/tile-structure.html)

## Tile Software Dependencies

[Tile Generator](https://docs.pivotal.io/tiledev/2-10/tile-generator.html)

[PCF CLI](https://docs.pivotal.io/tiledev/2-10/pcf-command.html)

## Development Workflow

### Environment setup

- Refer to [this guide](https://github.com/signalfx/signalfx-agent/tree/main/pkg/monitors/cloudfoundry)
on how to setup the Tanzu environment and local authentication information.
- A local file in your working directory is required for PCF commands to work. This file must be named `metadata`.
Example contents:
```yaml
---
# Values are found in the hammer file from Self Service's "ops_manager" key
opsmgr:
  url: https://pcf.<TAS environment name>.cf-app.com
  username: pivotalcf
  password: plain_text_password 
```

### CLI Commands
```shell
$ ./make-latest-tile
# Alternatively, if just making changes to the tile config without touching the BOSH release, you can just run "tile build"
# instead of the whole make-latest-tile script

$ pcf import product/splunk-opentelemetry-collector-<TILE_VERSION>.pivotal
$ pcf install splunk-opentelemetry-collector <TILE_VERSION>

# Optional: Use a configuration file to set tile variables instead of manually in the Ops Manager UI
$ pcf configure splunk-opentelemetry-collector tile_config.yaml
```

Sample `tile_config.yaml` file contents:
```yaml
---
cloudfoundry_rlp_gateway_shard_id: "opentelemetry"
cloudfoundry_rlp_gateway_tls_insecure_skip_verify: true
# Note: UAA credentials are from the UAA user created in the Tanzu service setup referenced
cloudfoundry_uaa_password: { 'secret': <UAA_PASSWORD> }
cloudfoundry_uaa_username: <UAA_USERNAME>
cloudfoundry_uaa_tls_insecure_skip_verify: true
splunk_access_token: <ACCESS_TOKEN>
splunk_realm: us0
otel_proxy_http: ""
otel_proxy_https: ""
otel_proxy_no: ""
splunk_api_url: ""
splunk_ingest_url: ""
```

Once tile is installed and configured you can go to the Ops Manager in your browser. Confirm your
`Splunk Opentelemetry Collector` tile is there, green, and the correct version. Select `Review Pending Changes` ->
Check box for staged changes on your tile -> `APPLY CHANGES`

Once changes are successfully applied, you should see data populating the charts in the Splunk Cloud Observability Suite

### Debugging Common Issues

1. Tile fails to start. The most common issue is UAA authentication errors or an incorrect Splunk access token.
   - Check logs found under the `CHANGE LOGS` header within the Tanzu Ops Manager.
   - Check logs for the tile itself
     - Click into tile
     - Check `Logs` tab, download if possible.
     - If no logs are found, go to `Status` tab, and click the download button in the `Logs` column. You can download
     the logs from the `Logs` tab once the download is finished.

2. Tile shows up as running, but charts aren't populating data. This is most likely a metric naming mismatch. TAS v3.0+
is currently unsupported due to metric name format changes.
    - Check logs to make sure no errors are showing up
    - Check metrics manually coming from Tanzu to see if they match charts.
      - Follow steps to
      [Access Metrics Using the Firehose PLugin](https://docs.pivotal.io/application-service/2-13/loggregator/data-sources.html#cf-nozzle)
        - You can use the `hammer` login command instead of `cf login`
        - Example metric filter command:
        ```cf nozzle -no-filter | grep -i mem | grep -i percent```
        - Compare output of CLI metric filter with chart variables in the Cloud foundry dashboard.