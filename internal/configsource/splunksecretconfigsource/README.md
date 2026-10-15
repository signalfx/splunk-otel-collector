# Splunk Secret Config Source

| Status | |
| --- | --- |
| Stability | Alpha |
| Distributions | [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector) |

Use the `splunk_secret` config source to retrieve secrets stored in Splunk's local
[secret storage](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/secretstorage/secretstoragerest)
(the `storage/passwords` REST endpoint served by splunkd on the management port,
typically `8089`) and inject them into your collector configuration.

This is primarily intended for use when the collector is launched as a Splunk TA
modular input, where the `SPLUNK_MANAGEMENT_URI` and `SPLUNK_SESSION_KEY` environment
variables are automatically set by the [TA launcher shim](../../../internal/modularinput/ta_launcher.go).
`endpoint` and `session_key` default to these environment variables, so the config
source works out-of-the-box when running as a TA modular input.

## Configuration

Under `config_sources:`, use `splunk_secret:` or `splunk_secret/<name>:` to create a
Splunk secret config source. `endpoint` and `session_key` are required unless their
respective environment variables are set. The TA launcher sets both variables, so
an empty `splunk_secret:` entry is sufficient when running as a TA modular input.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `endpoint` | Unless `SPLUNK_MANAGEMENT_URI` is set | `SPLUNK_MANAGEMENT_URI` environment variable | The splunkd management URI, such as `https://127.0.0.1:8089`. |
| `session_key` | Unless `SPLUNK_SESSION_KEY` is set | `SPLUNK_SESSION_KEY` environment variable | Authenticates requests to the `storage/passwords` endpoint. |
| `app` | No | `"-"` (all apps) | Limits the lookup to a Splunk app namespace; use it to disambiguate credentials with the same realm and name. |
| `user` | No | `"-"` (all users) | Limits the lookup to a Splunk user namespace; use it to disambiguate credentials with the same realm and name. |
| `insecure_skip_verify` | No | Unset | Set `true` to skip TLS certificate verification. When unset, verification is skipped only for literal loopback HTTPS endpoints; set `false` to require verification there. |
| `timeout` | No | `10s` | Maximum time to wait for a response from splunkd. Must not be negative. |

### Example configuration

When running as a TA modular input:

```yaml
config_sources:
  splunk_secret:
```

Outside a TA modular input, provide credentials explicitly:

```yaml
config_sources:
  splunk_secret:
    endpoint: https://splunk.example.com:8089
    session_key: ${env:SPLUNK_SESSION_KEY}
    app: search
    user: admin
```

## Selector syntax

Secrets are identified by a `<realm>:<name>` selector, matching the `realm` and
`name` fields used when the secret was created via the `storage/passwords`
endpoint. If the secret was stored without a realm, omit the realm segment but
keep the leading colon, e.g. `${splunk_secret::myuser}`.

A `<realm>:<name>` pair is only guaranteed unique within the Splunk user/app
namespace it was created in. If a selector matches credentials owned by more
than one user/app namespace, the lookup fails with an ambiguous selector error;
set the `user` and/or `app` config fields to disambiguate.

## Using a secret in a receiver

Configuring the [OpenTelemetry PostgreSQL receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/d975955280b0345300af84e7a029004541fb89ab/receiver/postgresqlreceiver#postgresql-receiver)
to monitor the PostgreSQL instance installed in Splunk Enterprise search heads (Linux only),
can be configured by creating the receiver as shown below and adding it to the metrics and
log pipelines of the collector:

```yaml
receivers:
  postgresql:
    username: postgres_admin
    password: ${splunk_secret:postgres:postgres_admin} # This secret is automatically set during Splunk Enterprise installation
```
