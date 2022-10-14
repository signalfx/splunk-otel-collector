# Oracle DB receiver

| Status                   |                            |
| ------------------------ |----------------------------|
| Stability                | [in-development]           |
| Supported pipeline types | metrics                    |

This receiver collects metrics from an Oracle Database.

The receiver connects to a database host and performs periodically queries.

Supported pipeline types: metrics

## Getting Started

The following settings are required:

- `datasource`: Oracle database connection string. Refer to Oracle Go Driver go_ora documentation for full connection string options.

Example:

```yaml
receivers:
  oracledb:
    datasource: "oracle://otel:password@localhost:51521/XE"
```

## Permissions

Depending on which metrics you collect, you will need to assign those permissions to the database user:
```
GRANT SELECT ON V_$SESSION TO <username>;
GRANT SELECT ON V_$SYSSTAT TO <username>;
GRANT SELECT ON V_$RESOURCE_LIMIT TO <username>;
GRANT SELECT ON DBA_TABLESPACES TO <username>;
GRANT SELECT ON DBA_DATA_FILES TO <username>;
```

## Enabling metrics.

See [documentation.md]. 

You can enable or disable selective metrics.

Example:

```yaml
receivers:
  oracledb:
    datasource: "oracle://otel:password@localhost:51521/XE"
    metrics:
      oracledb.query.cpu_time:
        enabled: false
      oracledb.query.physical_read_requests:
        enabled: true
```
