# Development Guide

## VMware Tanzu Tile Documentation

[Tanzu Tile Introduction](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/tile-basics.html)

[How Tiles Work](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/tile-structure.html)

## Tile Software Dependencies

[Tile Generator](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/tile-generator.html) -
Note that MacOS support was dropped for this tool, so an older version must be downloaded for darwin development.
Version `14.0.6-dev.1` has been confirmed to be work.

[PCF CLI](https://docs.vmware.com/en/VMware-Tanzu-Operations-Manager/3.0/tile-dev-guide/pcf-command.html)

## Development Workflow

### Install Required CLI tools

Refer to the [install_cli_dependencies script](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/tile/scripts#install_cli_dependenciessh)
for information on how to install all required CLI tools locally.

### Tanzu Environment setup

Refer to [this guide](https://github.com/signalfx/signalfx-agent/tree/main/pkg/monitors/cloudfoundry)
on how to setup the Tanzu environment and local authentication information. The guide has also been copied to a
[script](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/tile/scripts#setup_tanzush) that can be used to prepare the Tanzu environment for deploying the tile.


### Create a Tanzu Tile

Create a Tanzu Tile with the latest Splunk OpenTelemetry Collector

```shell
$ ./make-latest-tile
# Alternatively, if just making changes to the tile config without touching the BOSH release, you can just run "tile build"
# instead of the whole make-latest-tile script
```

### Configuring the Tile in the Tanzu Environment

The Tanzu Tile created must be imported, configured, and deployed in your Tanzu environment for testing. The 
import and configuration process can be done either via the CLI or Tanzu Ops Manager UI, as described below.
The deployment of the tile is done via the Tanzu Ops Manager UI.

#### CLI Configuration

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
```shell
$ pcf import product/splunk-otel-collector-<TILE_VERSION>.pivotal
$ pcf install splunk-otel-collector <TILE_VERSION>

# Optional: Use a configuration file to set tile variables instead of manually in the Ops Manager UI
$ pcf configure splunk-otel-collector tile_config.yaml
```

Sample `tile_config.yaml` file contents:
```yaml
---
cloudfoundry_rlp_gateway_tls_insecure_skip_verify: true
# Note: UAA credentials are from the UAA user created in the Tanzu service setup referenced
cloudfoundry_uaa_password: { 'secret': <UAA_PASSWORD> }
cloudfoundry_uaa_username: <UAA_USERNAME>
cloudfoundry_uaa_tls_insecure_skip_verify: true
splunk_access_token: <ACCESS_TOKEN>
splunk_realm: us0
otel_proxy_http: ""
otel_proxy_https: ""
otel_proxy_exclusions: ""
splunk_api_url: ""
splunk_ingest_url: ""
```

Once tile is installed and configured you can go to the Ops Manager in your browser. Confirm your
`Splunk Opentelemetry Collector` tile is there, green, and the correct version. Select `Review Pending Changes` ->
Check box for staged changes on your tile -> `APPLY CHANGES`

#### Ops Manager Configuration

- Browse to the Tanzu Ops Manager that you've created in the self service environment.
- Login using credentials provided in self service center.
- Upload Tanzu Tile
  - Click`IMPORT A PRODUCT` -> select the created Tanzu Tile.
  - The Tile will show up in the left window pane, simply click `+` next to it.
  - The Tile will be shown on the Installation Dashboard as not being configured properly.
- Configure Tanzu Tile
  - Assign AZs and Networks - Fill in required values, these do not have an impact on the Collector's deployment.
  - Nozzle Config - The two UAA arguments are required, use the values supplied by the [setup script]([./scripts/setup_tanzu.sh](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry/tile/scripts#setup_tanzush)). The default username is `my-v2-nozzle` and password is `password`.
  - Splunk Observability Cloud - These config options are directly mapped to the SignalFx's exporter options, so fill in values you use there.
  - Resource Config - No changes necessary.
  - Click Save after every page's changes.

### Deploying the Tanzu Tile

- Install the Tanzu Tile
  - Browse to Tanzu Ops Manager home page.
  - Ensure Tanzu Tile shows up with a green bar underneath (configuration is complete).
  - Select `REVIEW PENDING CHANGES`
  - Optional: Unselect `Small Footprint VMware Tanzu Application Service`. Unselecting this will speed up deployment time.
  - Select `APPLY CHANGES`

Once changes are successfully applied, you should see data populating the charts in the Splunk Cloud Observability Suite

### Debugging Common Issues

#### Run the collector locally

It may be helpful to test the collector in a local environment, outside of the Tanzu tile deployment.
Once your self service environment is up and running, and you've run the [setup script](./scripts/setup_tanzu.sh),
the collector can be configured with the
[Cloud Foundry receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/cloudfoundryreceiver)
pointed at your self service center environment.

#### Common issues

1. Tile fails to start. The most common issue is UAA authentication errors or an incorrect Splunk access token.
   - Check logs found under the `CHANGE LOGS` header within the Tanzu Ops Manager.
   - Check logs for the tile itself
     - Click into tile
     - Check `Logs` tab, download if possible.
     - If no logs are found, go to `Status` tab, and click the download button in the `Logs` column. You can download
     the logs from the `Logs` tab once the download is finished.

1. Tile starts successfully, but spamming logs with a message that looks like this:

    ```
    "kind": "receiver", "name": "cloudfoundry", "data_type": "metrics"} 2023-06-23T16:10:11.143Z info go-loggregator@v7.4.0+incompatible/rlp_gateway_client.go:87 unexpected status code 404: {"message":"resource not found","error":"not_found"} {"kind": "receiver", "name": "cloudfoundry", "data_type": "metrics"}
    ```
   This is likely a UAA client permissions issue. Here's how to check:

   ```
    $ uaac client get <USERNAME>
    scope: uaa.none
    client_id: admin
    resource_ids: none
    authorized_grant_types: client_credentials
    autoapprove:
    authorities: clients.read password.write clients.secret clients.write uaa.admin scim.write scim.read
    lastmodified: 1694730760000
   ```
   - Check the `authorities` key in the output, ensure it has the `logs.admin` authority listed. If the user does not have the proper authority, add it using this command:

   ```
    $ uaac client update <USERNAME> --authorities "<EXISTING-PERMISSIONS> logs.admin"
   ```
   - Where `<EXISTING-PERMISSIONS>` is the current contents of the scope section from the output from uaac contexts. Reference [here.](https://docs.cloudfoundry.org/uaa/uaa-user-management.html#changing-passwords)
   ```
    # Example update command sequence
    $ uaac contexts
    ...
    [29]*[https://uaa.sys.tas_env.cf-app.com]

    [0]*[identity]
    client_id: identity
    access_token: ...
    token_type: bearer
    expires_in: 43199
    scope: zones.read uaa.resource zones.write scim.zones uaa.admin cloud_controller.admin
    jti: ...

    ...
    $ uaac client update <USERNAME> --authorities "zones.read uaa.resource zones.write scim.zones uaa.admin cloud_controller.admin logs.admin"
   ```

   - Check again to ensure `logs.admin` is shown in the `authorities` list.

   ```
    $ uaac client get <USERNAME>
    scope: uaa.none
    client_id: admin
    resource_ids: none
    authorized_grant_types: client_credentials
    autoapprove:
    authorities: zones.read logs.admin uaa.resource zones.write scim.zones uaa.admin cloud_controller.admin
    lastmodified: 1694730760000
   ```

1. Tile shows up as running, but charts aren't populating with data, and the logs are being spammed with the following message:
    ```
    2024-03-06T14:47:28.447-0800	error	exporterhelper/queue_sender.go:97	Exporting failed. Dropping data.	{"kind": "exporter", "data_type": "metrics", "name": "signalfx", "error": "not retryable error: Permanent error: \"HTTP/2.0 401 Unauthorized\\r\\nContent-Length: 0\\r\\nDate: Wed, 06 Mar 2024 22:47:28 GMT\\r\\nServer: istio-envoy\\r\\nWww-Authenticate: Basic realm=\\\"Splunk\\\"\\r\\nX-Envoy-Upstream-Service-Time: 0\\r\\n\\r\\n\"", "dropped_items": 1}
    go.opentelemetry.io/collector/exporter/exporterhelper.newQueueSender.func1
    go.opentelemetry.io/collector/exporter@v0.95.0/exporterhelper/queue_sender.go:97
    go.opentelemetry.io/collector/exporter/internal/queue.(*boundedMemoryQueue[...]).Consume
    go.opentelemetry.io/collector/exporter@v0.95.0/internal/queue/bounded_memory_queue.go:57
    go.opentelemetry.io/collector/exporter/internal/queue.(*Consumers[...]).Start.func1
    go.opentelemetry.io/collector/exporter@v0.95.0/internal/queue/consumers.go:43
    ```
   This is an access token issue. Please use a valid value for the `Access token` text box on the tile's `Splunk Observability Cloud` page.

1. Tile shows up as running without error messages in the logs, but charts aren't populating data. This is most likely a metric naming mismatch. TAS v3.0+
is currently unsupported due to metric name format changes.
    - Check logs to make sure no errors are showing up
    - Check metrics manually coming from Tanzu to see if they match charts.
      - Follow steps to
      [Access Metrics Using the Firehose PLugin](https://techdocs.broadcom.com/us/en/vmware-tanzu/platform/elastic-application-runtime/10-2/eart/cli-plugin.html)
        - You can use the `hammer` login command instead of `cf login`
        - Example metric filter command:
        ```cf nozzle -no-filter | grep -i mem | grep -i percent```
        - Compare output of CLI metric filter with chart variables in the Cloud foundry dashboard.
