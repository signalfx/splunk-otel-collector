# Development Guide

## VMware Tanzu Tile Documentation

[Tanzu Tile Introduction](https://docs.vmware.com/en/Tile-Developer-Guide/3.0/tile-dev-guide/tile-basics.html)

[How Tiles Work](https://docs.vmware.com/en/Tile-Developer-Guide/3.0/tile-dev-guide/tile-structure.html)

## Tile Software Dependencies

[Tile Generator](https://docs.vmware.com/en/Tile-Developer-Guide/3.0/tile-dev-guide/tile-generator.html)

[PCF CLI](https://docs.vmware.com/en/Tile-Developer-Guide/3.0/tile-dev-guide/pcf-command.html)

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

Once changes are successfully applied, you should see data populating the charts in the Splunk Cloud Observability Suite

### Debugging Common Issues

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
      [Access Metrics Using the Firehose PLugin](https://docs.vmware.com/en/VMware-Tanzu-Application-Service/5.0/tas-for-vms/cli-plugin.html)
        - You can use the `hammer` login command instead of `cf login`
        - Example metric filter command:
        ```cf nozzle -no-filter | grep -i mem | grep -i percent```
        - Compare output of CLI metric filter with chart variables in the Cloud foundry dashboard.