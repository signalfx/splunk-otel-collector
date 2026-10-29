# Splunk authentication extension

The `splunk_auth` extension authenticates inbound requests with HEC tokens and
attaches the token's disabled status and index settings to the request context.

The extension is available in `otelcol`. Configure a Splunk management endpoint
and a session key with permission to list HEC tokens, then select the extension
in the receiver's authentication settings:

```yaml
extensions:
  splunk_auth:
    endpoint: https://splunk.example.com:8089
    session_key: ${env:SPLUNK_SESSION_KEY}
    tls:
      ca_file: /etc/collector/splunk-ca.pem

receivers:
  otlp:
    protocols:
      grpc:
        auth:
          authenticator: splunk_auth
      http:
        auth:
          authenticator: splunk_auth

exporters:
  debug: {}

service:
  extensions: [splunk_auth]
  pipelines:
    logs:
      receivers: [otlp]
      exporters: [debug]
```

Clients send `Authorization: Splunk <HEC token>` over HTTP, or the equivalent
`authorization` gRPC metadata. Missing, unknown, stale, and disabled tokens are
rejected. The management session key is a separate credential from the HEC tokens.
It is redacted in configuration output and is fixed for the extension's lifetime;
refreshing HEC tokens does not renew that credential.

`endpoint` defaults to the `SPLUNK_MANAGEMENT_URI` environment variable, and
`session_key` defaults to `SPLUNK_SESSION_KEY`, matching the `splunk_secret`
config source. The modular input launcher supplies these environment variables
when running as a Splunk TA. Explicit YAML values override the defaults; when
an environment variable is unset or empty, its corresponding field must be
configured in YAML. `tls` accepts
the standard OpenTelemetry client TLS settings. Certificate verification is
enabled by default, including for loopback endpoints.

The extension authenticates requests and attaches index settings to their context;
it does not enforce index authorization or rewrite telemetry. The modular input
performs that additional authorization in its consumers.

Programmatic callers can use `NewFactoryWithTokenSource(source)` instead of
`NewFactory()`. Do not combine an injected source with `endpoint` or `session_key`.
Injected-source factories and `NewDefaultConfig()` do not read these environment
variables, since the source supplies its own management credentials.
Each `TokenSource.Load(ctx)` must
return a complete token listing, honor context cancellation, and return an error
if any part of retrieval fails. The source must support concurrent loads when
shared by multiple extension instances. Returned slices must remain stable while
the extension copies them. A source can retrieve tokens from Splunk REST, a file,
or another store without changing authentication or refresh scheduling.

`NewTokenSource(serverURI, sessionKey)` provides the Splunk REST
implementation. It retains the existing pagination and management endpoint TLS
behavior. Its session key is fixed for the source's lifetime; refreshing HEC
tokens does not renew that management credential.

The factory's configuration controls refresh behavior:

| Setting | Default | Meaning |
| --- | --- | --- |
| `refresh_interval` | `3m` | How often to fetch a new token snapshot. |
| `refresh_timeout` | `30s` | Deadline for an entire load, including all pages. |
| `max_staleness` | `10m` | Maximum age of a snapshot accepted for authentication. |

All durations must be positive, and `max_staleness` must be strictly greater
than `refresh_interval + refresh_timeout`; configuration validation rejects
values at or below that boundary. Snapshot age is measured from the beginning
of its load, so allow additional margin for retrieval time and scheduling delays.

`Start` loads the initial snapshot synchronously and fails if retrieval fails.
It then starts a single refresh worker. Call `Start` before starting receivers
and `Shutdown` after stopping them, including cleanup when receiver startup
fails. `Shutdown` cancels pending retrieval and waits for the worker to exit.

Successful loads replace the complete snapshot atomically. An empty listing
revokes all tokens. Failed loads preserve the last successful snapshot, but
authentication rejects it once `max_staleness` is reached. Retrieval continues,
and a successful load restores authentication. Requests already authenticated
retain their original authorization metadata; subsequent requests use the new
snapshot. Authentication never performs network requests.

`splunk-connect-for-otlp` injects the REST source and accepts the same three
settings in `inputs.conf`, using the defaults above when they are omitted. See
[the input configuration documentation](../../../../packaging/otlpinput/README.md#token-refresh)
for an example.
The previous static `tokens` configuration has been replaced by the injected
source and these refresh settings.
