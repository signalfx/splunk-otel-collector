# Splunk Secret Config Source (Alpha)

Use the `splunk_secret` config source to retrieve secrets stored in Splunk's local
[secret storage](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/secretstorage/secretstoragerest)
(the `storage/passwords` REST endpoint served by splunkd on the management port,
typically `8089`) and inject them into your collector configuration.

This is primarily intended for use when the collector is launched as a Splunk TA
modular input, where the `SPLUNK_MANAGEMENT_URI` and `SPLUNK_SESSION_KEY` environment
variables are automatically set by the [TA launcher shim](../../../pkg/modularinput/ta_launcher.go).
`endpoint` and `session_key` default to these environment variables, so the config
source works out-of-the-box when running as a TA modular input.

## Configuration

Under `config_sources:` use `splunk_secret:` or `splunk_secret/<name>:` to create a
Splunk secret config source. The following parameters are available:

```yaml
config_sources:
  splunk_secret:
    # endpoint is the splunkd management URI, e.g. "https://127.0.0.1:8089".
    # Defaults to the SPLUNK_MANAGEMENT_URI environment variable.
    endpoint: ${env:SPLUNK_MANAGEMENT_URI}
    # session_key authenticates requests to the storage/passwords endpoint.
    # Defaults to the SPLUNK_SESSION_KEY environment variable.
    session_key: ${env:SPLUNK_SESSION_KEY}
    # app scopes the storage/passwords lookup to a specific Splunk app namespace.
    # Set this to disambiguate credentials that collide across apps. Defaults to "-" (all apps).
    app: "-"
    # user scopes the storage/passwords lookup to a specific Splunk user namespace.
    # Set this to disambiguate credentials that collide across users. Defaults to "-" (all users).
    user: "-"
    # insecure_skip_verify controls whether the TLS certificate presented by splunkd
    # is verified. Defaults to false.
    insecure_skip_verify: false
    # timeout is the maximum amount of time to wait for a response from splunkd.
    # Defaults to 10s.
    timeout: 10s
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

## Example

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
