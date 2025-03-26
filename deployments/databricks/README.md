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
