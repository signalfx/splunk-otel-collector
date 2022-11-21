# Databricks Receiver (Alpha)

The Databricks Receiver uses the Databricks
[API](https://docs.databricks.com/dev-tools/api/latest/index.html)
to generate metrics about the operation of a Databricks instance.

Supported pipeline types: `metrics`

> :construction: This receiver is in **ALPHA**. Behavior, configuration fields, and metric data model are subject to change.

## Configuration

The following fields are required:

- `instance_name`: A string representing the name of the instance. This value gets set as a `databricks.instance.name` resource attribute.
- `endpoint`: The protocol (http or https), hostname, and port for the Databricks API, without a trailing slash.
- `token`: An [access token](https://docs.databricks.com/dev-tools/api/latest/authentication.html) to authenticate to the Databricks API. 

The following fields are optional:

- `collection_interval`: How often this receiver fetches information from the Databricks API.
Must be a string readable by [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Defaults to **30s**.
- `max_results`: The maximum number of items to return per API call. Defaults to **25** which is the maximum value.
If set explicitly, the API requires a value greater than 0 and less than or equal to 25.

### Example

```yaml
receivers:
  databricks:
    instance_name: my-instance
    endpoint: https://my.host
    token: abc123
    collection_interval: 60s
    max_results: 10
```

### Databricks Spark

To get the configuration parameters this receiver will need to get Spark metrics, run the following Scala notebook and note its output:

```
%scala
val orgId = spark.conf.get("spark.databricks.clusterUsageTags.clusterOwnerOrgId")
val apiEndpoint = dbutils.notebook.getContext.apiUrl.get
val uiPort = spark.conf.get("spark.ui.port")
```
