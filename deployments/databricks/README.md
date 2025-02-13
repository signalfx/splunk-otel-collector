# Databricks

## Overview

[Databricks](https://www.databricks.com/) is a data intelligence platform that can be used for
[data sharing](https://www.databricks.com/product/data-sharing),
[data engineering](https://www.databricks.com/product/data-engineering),
[artificial intelligence](https://www.databricks.com/product/artificial-intelligence),
[real-time streaming](https://www.databricks.com/product/data-streaming), and more.

The OpenTelemetry Collector can be used to observe the real-time state of a Databricks cluster to
ensure it's operating as expected and reduce mean time to repair (MTTR) in the case of downgraded performance.

## Deployment

### Overview

The Splunk distribution of the OpenTelemetry Collector can be deployed on a Databricks cluster using an
[init script](https://docs.databricks.com/en/init-scripts/index.html).
The Collector will run on every node in a Databricks cluster, gathering host and Apache Spark metrics.

Databricks recommends running [cluster-scoped](https://docs.databricks.com/en/init-scripts/cluster-scoped.html)
init scripts. This can be deployed as cluster-scoped or
[global](https://docs.databricks.com/en/init-scripts/global.html#).

### Configuration

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

1. `OTEL_VERSION` - Version of the Splunk distribution of the OpenTelemetry Collector to deploy. Default: `latest`
1. `SPLUNK_REALM` - [Splunk Observability Cloud realm](https://docs.splunk.com/observability/en/get-started/service-description.html#sd-regions)
to send data to. Default: `us0`
1. `SCRIPT_DIR` - Installation path for the Collector and its config on a Databricks node. Default: `/tmp/collector_download`

### How to Deploy

#### Deploy as a cluster-scoped init script

1. Set required environment variables in your Databricks environment.
1. Use the [deployment script](./deploy_collector.sh) and follow documentation for how to
[configure a cluster-scoped init script using the UI](https://docs.databricks.com/en/init-scripts/cluster-scoped.html#configure-a-cluster-scoped-init-script-using-the-ui)

#### Deploy as a global-scoped init script

1. Set required environment variables in your Databricks environment.
1. Use the deployment script and follow documentation for how to
[add a global init script using the UI](https://docs.databricks.com/en/init-scripts/global.html#add-a-global-init-script-using-the-ui).