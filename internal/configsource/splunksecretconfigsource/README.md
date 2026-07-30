# Splunk Secret Config Source (Alpha)

Use the `splunk_secret` config source to retrieve secrets stored in Splunk's local
[secret storage](https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/secretstorage/secretstoragerest)
(the `storage/passwords` REST endpoint served by splunkd on the management port,
typically `8089`) and inject them into your collector configuration.

This is primarily intended for use when the collector is launched as a Splunk TA
modular input, where the `SPLUNK_SERVER_URI` and `SPLUNK_SESSION_KEY` environment
variables are automatically set by the TA launcher shim (see
`pkg/modularinput/ta_launcher.go`) from the `server_uri` and `session_key` fields
Splunk provides on stdin. `endpoint` and `session_key` default to the values of
these environment variables, so the config source works out-of-the-box when
running as a TA modular input without any explicit configuration.

## Configuration

Under `config_sources:` use `splunk_secret:` or `splunk_secret/<name>:` to create a
Splunk secret config source. The following parameters are available:

```yaml
config_sources:
  splunk_secret:
    # endpoint is the splunkd management URI, e.g. "https://127.0.0.1:8089".
    # Defaults to the SPLUNK_SERVER_URI environment variable set by the TA
    # launcher shim when running as a TA modular input.
    endpoint: ${env:SPLUNK_SERVER_URI}
    # session_key authenticates requests to the storage/passwords endpoint.
    # Defaults to the SPLUNK_SESSION_KEY environment variable set by the TA
    # launcher shim when running as a TA modular input.
    session_key: ${env:SPLUNK_SESSION_KEY}
    # app scopes the storage/passwords lookup to a specific Splunk app namespace. A
    # <realm>:<name> credential is only guaranteed unique within the app it was created
    # in, so if multiple apps store a credential with the same realm and name, set this
    # to the owning app to disambiguate. Defaults to "-" (all apps).
    app: "-"
    # user scopes the storage/passwords lookup to a specific Splunk user namespace. A
    # <realm>:<name> credential can also be stored with user-level sharing, so if
    # multiple users store a credential with the same realm and name, set this to the
    # owning user to disambiguate. Defaults to "-" (all users).
    user: "-"
    # insecure_skip_verify controls whether the TLS certificate presented by splunkd
    # is verified. Defaults to false: certificate verification is enabled unless
    # explicitly disabled. Splunk's management port typically presents a
    # self-signed certificate, so set this to true only in trusted environments
    # or after having installed a trusted certificate on splunkd.
    insecure_skip_verify: false
    # timeout is the maximum amount of time to wait for a response from splunkd.
    # Defaults to 10s if not specified.
    timeout: 10s
```

## Selector syntax

Secrets are identified by a `<realm>:<name>` selector, matching the `realm` and
`name` fields used when the secret was created via the `storage/passwords`
endpoint. If the secret was stored without a realm, omit the realm segment but
keep the leading colon, e.g. `${splunk_secret::myuser}`.

A `<realm>:<name>` pair is only guaranteed unique within the Splunk user/app
namespace it was created in, so different users and/or apps can each store a
credential with the same realm and name. If a selector matches credentials
owned by more than one user/app namespace, the lookup fails with an ambiguous
selector error; set the `user` and/or `app` config fields to the owning
user/app to disambiguate.

```yaml
components:
  component_using_splunk_secret:
    api_key: ${splunk_secret:myrealm:myuser}
```

If multiple splunkd endpoints or credentials are needed, create different
instances of the config source, e.g.:

```yaml
config_sources:
  splunk_secret:
    endpoint: ${env:SPLUNK_SERVER_URI}
    session_key: ${env:SPLUNK_SESSION_KEY}

components:
  component_using_splunk_secret:
    api_key: ${splunk_secret:myrealm:myuser}
```
