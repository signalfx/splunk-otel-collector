# SQL Receiver (Alpha)

The SQL Receiver uses custom SQL queries to generate metrics from a database
connection.

> :construction: This receiver is in **ALPHA**. Behavior, configuration fields, and metric data model are subject to change.

## Configuration

The config supports three fields, all of which are required:

- `driver`: The name of the database driver: _postgres_, _mysql_, _snowflake_, _sqlserver_, or _hdb_ (SAP HANA).
- `datasource`: The datasourcename value passsed to [sql.Open](https://pkg.go.dev/database/sql#Open). e.g. `host=localhost port=5432 user=me password=s3cr3t sslmode=disable`
- `queries`: One or more queries, where a query is a sql statement and one or more metrics (details below).

### Queries

A _query_ consists of a sql statement and one or more _metrics_, where each metric consists of a
`metric_name`, a `value_column`, and `attribute_columns`. Each _metric_
will produce one OTel metric per database row returned from its associated sql query.

* `metric_name`(required): the name assigned to the OTel metric.
* `value_column`(required): the column name in the returned dataset used to set the value of the metric's datapoint. Must be an integer value.
* `attribute_columns`(optional): a list of column names in the returned dataset used to set attibutes on the datapoint.
* `is_monotonic`(optional): a boolean value for whether the value is monotonically increasing. If it is, the receiver will emit a sum type for this metric.

### Example

```yaml
receivers:
  sql:
    driver: postgres
    datasource: "host=localhost port=5432 user=postgres password=s3cr3t sslmode=disable"
    queries:
      - sql: "select count(*) as count, 42 as val from movie"
        metrics:
          - metric_name: movie.count
            value_column: "count"
          - metric_name: movie.val
            value_column: "val"
      - sql: "select count(*) as count, genre from movie group by genre"
        metrics:
          - metric_name: movie.genres
            value_column: "count"
            attribute_columns: [ "genre" ]
```