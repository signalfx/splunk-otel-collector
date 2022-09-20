# Oracle DB receiver

| Status                   |                            |
| ------------------------ |----------------------------|
| Stability                | [in-development]           |
| Supported pipeline types | metrics                    |
| Distributions            | [contrib]                  |

This receiver collects metrics from an Oracle Database.

The receiver connects to a database host and performs periodically queries.

Supported pipeline types: metrics

## Getting Started

The following settings are required:

- `username`: Oracle database account username
- `password`: Oracle database account password
- `endpoint`: Oracle database connection endpoint, of the form `host:port`
- `service_name`: Oracle database service name

Example:

```yaml
receivers:
  oracledb:
    username: otel
    password: password
    endpoint: localhost:51521
    service_name: XE
```

## Permissions

Depending on which metrics you collect, you will need to assign those permissions to the database user:
```
GRANT SELECT ON V_$SQLSTATS TO <username>;
GRANT SELECT ON V_$SESSION TO <username>;
GRANT SELECT ON V_$SESSTAT TO <username>;
GRANT SELECT ON V_$STATNAME TO <username>;
GRANT SELECT ON V_$SESSMETRIC TO <username>;
```

