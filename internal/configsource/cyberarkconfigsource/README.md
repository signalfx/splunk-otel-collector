# CyberArk Config Source (Alpha)

Use the CyberArk config source to retrieve credentials from
[CyberArk](https://www.cyberark.com/) and inject them into your collector configuration,
keeping secrets out of the collector config files.

This config source targets the **CyberArk Credential Provider (CP / AAM credential
provider agent)** installed on the collector host. It retrieves secrets by executing the
provider's bundled `CLIPasswordSDK` binary — there is no network endpoint to configure.
The CP agent maintains a local, auto-synced cache, so retrievals reflect the current
credential at process start.

> The Central Credential Provider (CCP) REST API is not yet supported; `retrieval_mode`
> defaults to `cp` and only `cp` is currently accepted. `ccp` is reserved for a future
> release.

## Configuration

Under `config_sources:` use `cyberark:` or `cyberark/<name>:` to create a CyberArk config
source. Each config source instance is pinned to a single CyberArk object.

```yaml
config_sources:
  cyberark:
    # retrieval_mode selects the CyberArk backend. Only "cp" (the local CLIPasswordSDK
    # binary) is currently supported. Defaults to "cp".
    retrieval_mode: cp
    # binary_path is the path to the CLIPasswordSDK executable (cp mode). It may be an
    # absolute path or a name resolvable on PATH. Defaults to "CLIPasswordSDK".
    binary_path: /opt/CARKaim/sdk/CLIPasswordSDK
    # app_id is the CyberArk Application ID authorized to retrieve the object. Required.
    app_id: collector-app
    # safe is the CyberArk safe holding the object. Required.
    safe: DBSecrets
    # folder is the folder within the safe. Optional; defaults to "Root".
    folder: Root
    # object is the name of the CyberArk object (account) to retrieve. Required.
    object: prod-db
    # auto_refresh enables a polling watcher that re-fetches the object and triggers a
    # collector config reload when the retrieved values change. Defaults to false.
    auto_refresh: false
    # poll_interval is how often the object is re-fetched when auto_refresh is true.
    # Defaults to 1 minute. Ignored when auto_refresh is false.
    poll_interval: 1m
```

## Selectors

A single retrieval fetches the account password plus a fixed set of pass properties and
object metadata. The reference selector picks which field to inject. An empty selector
returns the `Password`.

| Selector    | CyberArk attribute      |
|-------------|-------------------------|
| (empty)     | `Password`              |
| `Password`  | `Password`              |
| `UserName`  | `PassProps.UserName`    |
| `Address`   | `PassProps.Address`     |
| `Database`  | `PassProps.Database`    |
| `Port`      | `PassProps.Port`        |
| `Name`      | `Name`                  |
| `Safe`      | `Safe`                  |
| `Folder`    | `Folder`                |

```yaml
components:
  some_receiver:
    endpoint: ${cyberark:Address}:${cyberark:Port}
    username: ${cyberark:UserName}
    password: ${cyberark:Password}
```

If multiple objects are needed, create different instances of the config source:

```yaml
config_sources:
  cyberark/db:
    app_id: collector-app
    safe: DBSecrets
    object: prod-db
  cyberark/api:
    app_id: collector-app
    safe: APISecrets
    object: partner-api

components:
  db_receiver:
    username: ${cyberark/db:UserName}
    password: ${cyberark/db:Password}
  api_client:
    token: ${cyberark/api:Password}
```

## Credential rotation

CyberArk does not lease credentials. A credential object has a stable identity (safe +
object); its password *value* is rotated by the CyberArk CPM on a policy schedule or
on-demand. There is no lease, TTL, or renewal to react to — rotation just changes the value
of an object that continues to exist. This config source therefore does not implement any
lease/renewal handling; it deals with rotation by (optionally) re-reading the value.

### How a config source value reaches the collector

A config source resolves `${cyberark:...}` references **once**, when the collector loads (or
reloads) its configuration. The resolved value is a plain literal baked into each
component's config. A receiver never holds a live reference to CyberArk and cannot re-fetch
a credential on its own — the only way a new credential reaches a running collector is to
re-resolve the whole configuration. That has two consequences for rotation:

**1. Static mode (default, `auto_refresh: false`).** Values are fetched once and never
watched. If CyberArk rotates the password *after* the collector has started, the collector
keeps using the value it resolved at startup — it will not notice the change on its own, and
components that authenticate once at startup keep working while components that re-connect
later may begin to fail with the stale credential. The credential is refreshed only when the
collector process restarts (systemd, Kubernetes, or a component fatal error) and re-resolves
its config. Because the CP agent keeps a local, auto-synced cache, that restart always reads
the current value. This mode is the right default when rotation is infrequent and a restart
is an acceptable way to pick it up.

**2. Auto-refresh mode (`auto_refresh: true` + `poll_interval`).** The config source starts a
background watcher that re-runs the retrieval every `poll_interval`. When the retrieved
values change, it signals the collector to **reload its configuration**, which re-resolves
every `${cyberark:...}` reference to the new value and rebuilds the affected component graph
in-process — no process restart. Collector impact to be aware of:

- The reload is a full config re-resolution and component-graph restart, so affected
  pipelines briefly tear down and rebuild. This is the same reload mechanism the `vault`
  config source uses.
- Detection is bounded by `poll_interval`: a rotation is picked up at most one poll interval
  after it happens. Between rotation and the next poll, the old value is still in use.
- Each poll executes `CLIPasswordSDK`; pick a `poll_interval` that balances rotation latency
  against exec/CPU cost. It defaults to 1 minute.
- If a poll fails (e.g. the object is temporarily unavailable), the watcher signals a reload
  with the error so it surfaces at resolution time rather than being silently swallowed.

Choose static mode when a restart on rotation is fine; choose auto-refresh when credentials
rotate often enough that in-process refresh is worth the periodic reload.

## Parsing assumptions

The `cp` retriever requests a fixed, ordered list of attributes from `CLIPasswordSDK` using
its `-o` flag, joined with a rare delimiter (`@#@`) on a single output line, and parses the
result positionally. It assumes the SDK emits every requested field, in order, and that
unset pass properties still emit an empty positional value so the field count stays stable.
A field-count mismatch is treated as an error rather than silently misassigning values. A
configurable field list is a possible future extension.

## Future enhancements

The current implementation targets the CyberArk Credential Provider (CP) via
`CLIPasswordSDK` and retrieves a fixed set of fields. The following are planned:

- **CCP retrieval mode (`retrieval_mode: ccp`).** Add a Central Credential Provider
  backend that talks mTLS REST to the AIMWebService endpoint instead of shelling out to a
  local agent. This reuses the existing `retriever` seam, account addressing
  (`app_id`/`safe`/`object`), field/selector model, caching, and auto-refresh watcher — no
  changes to `source.go`. It only adds a new retriever plus CCP-specific config
  (`endpoint`, client TLS material).
- **Conjur config source.** CyberArk Conjur exposes genuine dynamic secrets with leases
  and TTLs and uses a different addressing and authentication model. Rather than force it
  into a `retrieval_mode` here, it belongs in a separate config source that can model
  lease-based renewal (closer to how the `vault` source handles dynamic secrets) instead of
  the rotation/polling model used for CP.
- **Configurable property retrieval.** Today the `cp` retriever requests a fixed, ordered
  list of attributes (`outputFields`). A future enhancement would let operators declare
  which object properties to fetch and expose, so custom safe/object properties beyond the
  built-in set become selectable.
