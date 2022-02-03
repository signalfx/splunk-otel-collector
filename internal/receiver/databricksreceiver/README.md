# Databricks Receiver (Alpha)

The Databricks Receiver uses the Databricks
[API](https://docs.databricks.com/dev-tools/api/latest/index.html)
to generate metrics about the operation of a Databricks instance.

Supported pipeline types: `metrics`

> :construction: This receiver is in **ALPHA**. Behavior, configuration fields, and metric data model are subject to change.

## Configuration

The following fields are required:

- `instance_name`: A string representing the name of the instance. This value gets set as a `databricks.instance.name` resource attribute.
- `base_url`: The protocol (http or https), hostname, and port for the Databricks API, without a trailing slash.
- `token`: An [access token](https://docs.databricks.com/dev-tools/api/latest/authentication.html) to authenticate to the Databricks API. 

The following field is optional:

- `collection_interval`: How often this receiver fetches information from the Databricks API.
Must be a string readable by [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Defaults to **30s**.

### Example

```yaml
receivers:
  databricks:
    instance_name: my-instance
    base_url: https://my.host
    token: abc123
    collection_interval: 10s
```
