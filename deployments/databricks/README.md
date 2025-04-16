# Databricks

## Overview

[Databricks](https://www.databricks.com/) is a data intelligence platform that can be used for
[data sharing](https://www.databricks.com/product/data-sharing),
[data engineering](https://www.databricks.com/solutions/data-engineering),
[artificial intelligence](https://www.databricks.com/product/artificial-intelligence),
[real-time streaming](https://www.databricks.com/product/data-streaming), and more.

The OpenTelemetry Collector can be used to observe the real-time state of a Databricks cluster to
ensure it's operating as expected and reduce mean time to repair (MTTR) in the case of downgraded performance.
This functionality is only relevant for the classic architecture of Databricks, it will not work for
serverless.

## Deployment

### Overview

The Splunk distribution of the OpenTelemetry Collector can be deployed on a Databricks cluster using an
[init script](https://docs.databricks.com/en/init-scripts/index.html) on start up, or by directly running
the script on each node in an existing running Databricks cluster.
The Collector will run on every node in a Databricks cluster, gathering host and Apache Spark metrics.

### Init script

Databricks recommends running [cluster-scoped](https://docs.databricks.com/en/init-scripts/cluster-scoped.html)
init scripts. This can be deployed as cluster-scoped or
[global](https://docs.databricks.com/en/init-scripts/global.html#).

#### Configuration

The init script uses the following environment variables. The variables can be set
via
[Databricks init script environment variables](https://docs.databricks.com/en/init-scripts/environment-variables.html),
or by directly setting the values in the init script itself.

#### Required Environment Variables

1. `SPLUNK_ACCESS_TOKEN` - Set to your  [Splunk Observability Cloud access token](https://docs.splunk.com/observability/en/admin/authentication/authentication-tokens/org-tokens.html) 
1. `DATABRICKS_ACCESS_TOKEN` - Set to your [Databricks personal access token](https://docs.databricks.com/en/dev-tools/auth/pat.html)
1. `DATABRICKS_CLUSTER_HOSTNAME` - Hostname of the [Databricks compute resource](https://docs.databricks.com/en/integrations/compute-details.html).
(Use the "Server Hostname")

#### Optional Environment Variables

1. `SPLUNK_OTEL_VERSION` - Version of the Splunk distribution of the OpenTelemetry Collector to deploy. Default: `latest`
1. `SPLUNK_REALM` - [Splunk Observability Cloud realm](https://docs.splunk.com/observability/en/get-started/service-description.html#sd-regions)
to send data to. Default: `us0`
1. `SCRIPT_DIR` - Installation path for the Collector and its config on a Databricks node. Default: `/tmp/collector_download`

#### How to Deploy

##### Deploy as a cluster-scoped init script

1. Set required environment variables in your Databricks environment.
1. Use the [deployment script](./deploy_collector.sh) and follow documentation for how to
[configure a cluster-scoped init script using the UI](https://docs.databricks.com/en/init-scripts/cluster-scoped.html#configure-a-cluster-scoped-init-script-using-the-ui)

##### Deploy as a global-scoped init script

1. Set required environment variables in your Databricks environment.
1. Use the deployment script and follow documentation for how to
[add a global init script using the UI](https://docs.databricks.com/en/init-scripts/global.html#add-a-global-init-script-using-the-ui).

### Standalone script

For long running clusters, restarting the whole cluster to run an init script on each
node may not be a feasible option. In this case, the deployment script can be run on
each node manually.

#### Configuration

The required and optional environment variables outlined in the init script section remain
the same, but more variables are required.

##### Required Environment Variables

These environment variables are required **in addition** to what's required for init scripts.
All required environment variables must be set on every node that runs the deployment script.

1. `DB_IS_DRIVER` - whether the script is running on a driver node. (boolean)
1. `DB_CLUSTER_NAME` - the name of the cluster the script is executing on. (string)
1. `DB_CLUSTER_ID` - the ID of the cluster on which the script is running. See the [Clusters API](https://docs.databricks.com/api/workspace/clusters). (string)

#### How to deploy

The Databricks cluster provides a web terminal on the driver node. This is a BASH shell
which can then be accessed to deploy the script.

**Note: Investigation is ongoing to determine how to deploy the script on non-driver nodes.**

## Debugging the Init Script

From testing, the init script may fail for a variety of reasons, this section is meant
to help users investigate the root cause of failing to get metrics from a Databricks cluster.

### Init script setup

As a first step of investigation, please ensure the init script has been properly configured.

1. Ensure all required environment variables are set in the script.
1. Ensure all required environment variables are set in the cluster configuration.
1. Ensure the cluster is configured to run the init script on startup.

### Situation 1: Cluster fails to start due to init script failure

- Enable [init script logging](https://learn.microsoft.com/en-us/azure/databricks/init-scripts/logs)
- Read through logs to see if any relevant information can be found

### Situation 2: Cluster is running but no data is seen in charts (enabled web terminal)

#### Pre-requisites

- Enable the [web terminal](https://learn.microsoft.com/en-us/azure/databricks/admin/clusters/web-terminal)

#### Investigate

1. Access the [web terminal](https://learn.microsoft.com/en-us/azure/databricks/compute/web-terminal)

1. Ensure the `splunk_otel_collector.service` is running

    ```bash
    $ systemctl # Check output for the service
    ```

1. Check contents of the Collector's configuration file

   ```bash
    $ cat /tmp/collector_download/config.yaml # This is the default location unless changed by user.
    ```

1. Check syslogs for possible errors coming from the Collector

    ```bash
    $ tail -n 50 /var/log/syslog
    ```

1. If the service is running, the configuration looks right, and nothing looks concerning from
the syslogs, check the SignalFx backend for metrics to see if it's possibly a dashboard
issue. Note that at the time of writing OOTB content has not been updated for OTel metrics.
There is currently no OOTB content for Databricks, and the Apache Spark dashboard is
for Smart Agent metrics. The only charts that show data are a subset of host metric
charts.

1. Confirm metric time series (MTS) limits are not being hit for the organization.
   - [MTS default limits per product](https://docs.splunk.com/observability/en/admin/references/per-product-limits.html#mts-limits-per-product)
   - [Access organization metrics](https://docs.splunk.com/observability/en/admin/org-metrics.html#org-metrics)

### Situation 3: Cluster is running but no data is seen in charts (disabled web terminal)

1. Check the SignalFx backend for metrics to see if it's possibly a dashboard issue.
Note that at the time of writing OOTB content has not been updated for OTel metrics.
There is currently no OOTB content for Databricks, and the Apache Spark dashboard is
for Smart Agent metrics. The only charts that show data are a subset of host metric
charts.

1. Confirm metric time series (MTS) limits are not being hit for the organization.
   - [MTS default limits per product](https://docs.splunk.com/observability/en/admin/references/per-product-limits.html#mts-limits-per-product)
   - [Access organization metrics](https://docs.splunk.com/observability/en/admin/org-metrics.html#org-metrics)
