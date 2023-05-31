# Databricks Receiver

The Databricks Receiver uses the Databricks
[API](https://docs.databricks.com/dev-tools/api/latest/index.html)
to generate metrics about the operation of a Databricks instance.

In addition to generating metrics from the Databricks API, it also generates metrics from the Spark subsystem running in a Databricks instance.

Supported pipeline types: `metrics`

> :construction: This receiver is in **DEVELOPMENT**. Behavior, configuration fields, and metric data model are subject to change.

## Configuration

The following fields are required:

- `instance_name`: A string representing the name of the instance. This value gets set as a `databricks.instance.name` resource attribute.
- `endpoint`: The URL containing a protocol (http or https), hostname, and (optional) port of the Databricks API, without a trailing slash.
- `token`: An [access token](https://docs.databricks.com/dev-tools/auth.html#databricks-personal-access-tokens) to authenticate to the Databricks API. 
- `spark_org_id`: The Spark Org ID. See the Spark Subsystem Configuration section below for how to get this value.
- `spark_endpoint`: The URL containing a protocol (http or https), hostname, and (optional) port of the Spark API. See the Spark Subsystem Configuration section below for how to get this value.
- `spark_ui_port`: A number representing the Spark UI Port (typically 40001). See the Spark Subsystem Configuration section below for how to get this value.

The following fields are optional:

- `collection_interval`: How often this receiver fetches information from the Databricks API.
Must be a string readable by [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Defaults to **30s**.
- `max_results`: The maximum number of items to return per API call. Defaults to **25** which is the maximum value.
If set explicitly, the API requires a value greater than 0 and less than or equal to 25.

### Example

```yaml
databricks:
  instance_name: my-instance
  endpoint: https://dbr.example.net
  token: abc123
  spark_org_id: 1234567890
  spark_endpoint: https://spark.example.net
  spark_ui_port: 40001
  collection_interval: 10s
```

### Spark Subsystem Configuration

To get the configuration parameters this receiver will need to get Spark metrics, run the following Scala notebook and copy its output values into your config:

```
%scala
val sparkOrgId = spark.conf.get("spark.databricks.clusterUsageTags.clusterOwnerOrgId")
val sparkEndpoint = dbutils.notebook.getContext.apiUrl.get
val sparkUiPort = spark.conf.get("spark.ui.port")
```
