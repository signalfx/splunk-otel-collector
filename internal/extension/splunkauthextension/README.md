# Splunk authentication extension

The `splunkauth` extension authenticates inbound requests with HEC tokens and
attaches the token's disabled status and index settings to the request context.

Construct the extension factory with a `TokenSource`. Each `Load(ctx)` must
return a complete token listing, honor context cancellation, and return an error
if any part of retrieval fails. The source must support concurrent loads when
shared by multiple extension instances. Returned slices must remain stable while
the extension copies them. A source can retrieve tokens from Splunk REST, a file,
or another store without changing authentication or refresh scheduling.

`internal/auth.NewTokenSource(serverURI, sessionKey)` provides the Splunk REST
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
[the input configuration documentation](../../../packaging/otlpinput/README.md#token-refresh)
for an example.
The previous static `tokens` configuration has been replaced by the injected
source and these refresh settings.
