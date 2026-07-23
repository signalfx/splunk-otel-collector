# Splunk Platform Log and Metrics Collection

> New to this feature? Start with the [quickstart guide](getting-started/linux-splunk-platform.md).

## Contents

- [Enabling Optional Receivers](#enabling-optional-receivers)
- [Common Scenarios](#common-scenarios)
  - [Scenario 1: Enable the `/etc` config file receiver](#scenario-1-enable-the-etc-config-file-receiver)
  - [Scenario 2: Add or remove paths from a receiver](#scenario-2-add-or-remove-paths-from-a-receiver)
  - [Scenario 3: Narrow down what `file_log/etc` collects](#scenario-3-narrow-down-what-file_logetc-collects)
  - [Scenario 4: Add multiline support to a receiver](#scenario-4-add-multiline-support-to-a-receiver)
  - [Scenario 5: Run alongside Splunk Observability Cloud agent](#scenario-5-run-alongside-splunk-observability-cloud-agent)
- [Best Practices](#best-practices)
- [Sending Queue and Backpressure](#sending-queue-and-backpressure)
- [Metrics](#metrics)
- [Known Limitations, Constraints, and Troubleshooting](#known-limitations-constraints-and-troubleshooting)
  - [Limitations](#limitations)
  - [Verifying Ingestion](#verifying-ingestion)
  - [Troubleshooting](#troubleshooting)

---

## Enabling Optional Receivers

Edit `/etc/otel/collector/splunk_logs_config_linux.yaml` and uncomment receivers in the `service.pipelines.logs/hec.receivers` section:

```yaml
service:
  pipelines:
    logs/hec:
      receivers:
        - file_log/varlog       # enabled by default
        #- file_log/etc         # uncomment to enable
        #- file_log/bash_history # uncomment to enable
        #- journald             # uncomment to enable
        #- file_log/nginx-access
        #- file_log/mysql-error
        # ... (see full list in the config file)
```

After editing, restart the service:

```sh
sudo systemctl restart splunk-otel-collector
```

---

## Common Scenarios

### Scenario 1: Enable the `/etc` config file receiver

The `file_log/etc` receiver is defined in the config but disabled by default. Uncomment it in the pipeline to start collecting config files from `/etc`:

```yaml
service:
  pipelines:
    logs/hec:
      receivers:
        - file_log/varlog
        - file_log/etc      # uncomment this line
```

`file_log/etc` is intentionally designed for a one-time read — it has no `storage` set, so it re-reads matching files from the beginning on every collector restart. This is useful for capturing the state of config files rather than tailing them continuously.

### Scenario 2: Add or remove paths from a receiver

Every `file_log` receiver has an `include` list of glob patterns controlling which files it tails. You can add new paths or remove ones you don't need by editing that list directly.

For example, to also collect logs from a custom application writing to `/opt/myapp/logs/`:

```yaml
file_log/varlog:
  include:
    - /var/log/*.log
    - /var/log/*log
    - /var/log/*messages*
    - /var/log/*secure*
    - /var/log/*auth*
    - /var/log/*mesg
    - /var/log/*cron
    - /var/log/*acpid
    - /var/log/*.out
    - /var/adm/*.log
    - /var/adm/*messages*
    - /opt/myapp/logs/*.log   # added
  # ... rest of existing config unchanged ...
```

You can also use `exclude` to skip specific files within a matched pattern:

```yaml
file_log/varlog:
  include:
    - /var/log/*.log
  exclude:
    - /var/log/noisy-app.log
```

See the [filelog receiver documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) for the full glob syntax and all available options.

### Scenario 3: Narrow down what `file_log/etc` collects

By default `file_log/etc` matches a broad set of extensions under `/etc`. To restrict it to specific files, edit the `include` list directly in the config:

```yaml
file_log/etc:
  include:
    - /etc/*.conf
    - /etc/*.cfg
    # remove extensions you don't need
```

After editing, restart the service.

### Scenario 4: Add multiline support to a receiver

Some log formats span multiple lines (e.g. Java stack traces, multi-line app logs). You can add a `multiline` block to any `file_log` receiver. See the [filelog receiver documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) and the [multiline configuration reference](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/operators/recombine.md) for the full set of options.

Example — configure `file_log/varlog` to treat lines starting with a timestamp as the beginning of a new event:

```yaml
file_log/varlog:
  include:
    - /var/log/*.log
    # ... existing include list ...
  multiline:
    line_start_pattern: '^\d{4}-\d{2}-\d{2}'  # lines starting with a date begin a new event
  # ... rest of existing config ...
```

### Scenario 5: Run alongside Splunk Observability Cloud agent

Load both configs using `--config` flags and the `confmap.enableMergeAppendOption` feature gate. The installer handles this automatically when both an access token and a platform URL are provided:

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh | sudo sh - \
  --realm us0 \
  --splunk-platform-token "YOUR_HEC_TOKEN" \
  --splunk-platform-url   "https://splunk.example.com:8088/services/collector" \
  -- YOUR_O11Y_ACCESS_TOKEN
```

The two configs merge cleanly — `agent_config.yaml` owns the o11y pipelines and `splunk_logs_config_linux.yaml` adds the `logs/hec` pipeline. They share the `health_check` and `zpages` extensions.

---

## Best Practices

### Always restart after config changes

The collector does not hot-reload config. After any edit to `splunk_logs_config_linux.yaml`, restart the service:

```sh
sudo systemctl restart splunk-otel-collector
```

### Validate config before restarting

Use `otelcol validate` to catch errors before they take the service down. The config references environment variables, so they must be set for validation to succeed — source them from the env file:

```sh
sudo env $(grep -v '^#\|^OTELCOL_OPTIONS' /etc/otel/collector/splunk-otel-collector.conf | xargs) \
  /usr/bin/otelcol validate --config /etc/otel/collector/splunk_logs_config_linux.yaml
```

### Don't enable journald without group membership

The `journald` receiver will silently fail if the service user is not in the `systemd-journal` group. Always run the following before enabling it:

```sh
sudo usermod -aG systemd-journal splunk-otel-collector
sudo systemctl restart splunk-otel-collector
```

---

## Sending Queue and Backpressure

The `splunk_hec/logs` exporter ships with a pre-tuned `sending_queue` configuration. This section explains the key settings and when you might need to adjust them.

### `block_on_overflow: true` — do not remove this

This is the most important setting in the queue configuration. When the sending queue fills (e.g. Splunk is slow or unreachable), `block_on_overflow: true` causes the filelog receiver to pause reading instead of dropping records. The unread lines accumulate on disk — the file itself acts as a durable buffer — and reading resumes when the queue drains.

Without this setting, a full queue causes records to be **silently and permanently dropped**. File offsets are advanced past the lost lines and they can never be recovered.

### Tune `queue_size` for your workload

The right `queue_size` depends on the number of files being tailed, log throughput, and HEC latency. The default value provides a reasonable starting point for most deployments.

A larger queue gives more headroom before `block_on_overflow` engages — useful when tailing many files with bursty write patterns. However, every queued record is a Go object in memory, so very large values (100k+) increase garbage collection pressure and can slow export throughput.

By default, the sending queue is in-memory and `wait_for_result` is disabled. This means if the collector restarts or crashes, **everything currently held in the queue is lost**. With this in mind, keeping `queue_size` smaller limits how much data is at risk — the filelog receiver's checkpoints ensure that unread data remains on disk and will be picked up after restart, so the exposure is limited to whatever was already queued but not yet delivered. If you need stronger guarantees, enable persistent queue storage as described below.

### `sending_queue.batch` — only for high-latency networks

The `batch` setting under `sending_queue` accumulates records before sending to reduce the number of HTTP round-trips. This helps when HEC round-trip latency is high (>100ms, e.g. cross-region). On low-latency networks (local or same-region) it adds wall-clock delay without improving throughput. The default config enables batching with conservative settings — if you observe high latency to your HEC endpoint, increasing `min_size` may help.

### Delivery guarantee: at-least-once

The HEC protocol has no idempotency mechanism. If a network connection drops mid-send, the exporter retries the batch — if Splunk had already indexed it, the events appear twice. Duplicates are expected and bounded by `num_consumers × batch_size`. In testing, the duplicate rate was ~0.2% during a network disruption event.

### `wait_for_result` — synchronous delivery

By default `wait_for_result` is disabled — the filelog receiver hands records to the queue and continues reading without waiting for HEC to acknowledge them. This maximises throughput but means in-flight records can be lost if the collector crashes.

Setting `wait_for_result: true` makes the filelog receiver block until each batch is acknowledged by Splunk before advancing its file offset. This limits crash loss to at most one in-flight batch, at the cost of reduced throughput (no parallelism between reading and sending). Use this only if your log volume is low enough that the throughput reduction is acceptable.

```yaml
exporters:
  splunk_hec/logs:
    wait_for_result: true
    sending_queue:
      enabled: true
      block_on_overflow: true
```

### Persistent queue for crash recovery

By default the sending queue is in-memory. If the collector process crashes or is restarted while records are queued (e.g. during a Splunk outage), those records are lost. To recover them after restart, enable persistent queue storage:

```yaml
extensions:
  file_storage/queue:
    directory: /var/lib/otelcol/queue

exporters:
  splunk_hec/logs:
    sending_queue:
      storage: file_storage/queue

service:
  extensions: [health_check, zpages, file_storage/filelogs, file_storage/queue]
```

With persistent storage enabled, the queue is written to disk on shutdown and resumed on restart. Note that 503 responses from Splunk during catch-up drain are expected and handled automatically by `retry_on_failure` — no records are dropped. Be aware that writing the queue to disk adds I/O overhead on every enqueue and dequeue operation, which reduces throughput compared to the in-memory default.

---

## Metrics

### Editing the host metrics receiver

The `host_metrics` receiver in `splunk_metrics_config_linux.yaml` collects a standard set of system scrapers by default. You can adjust which scrapers are active by editing the `scrapers` block:

```yaml
receivers:
  host_metrics:
    collection_interval: 10s
    scrapers:
      cpu:
      disk:
      filesystem:
      memory:
      network:
      load:
      paging:
      processes:
```

To change the collection interval, set `collection_interval` to the desired value (e.g. `30s`, `1m`).

### Enabling additional metrics disabled by default

Some scrapers are commented out because they are noisy, high-cardinality, or require elevated permissions. To enable them, uncomment the relevant scraper:

```yaml
receivers:
  host_metrics:
    collection_interval: 10s
    scrapers:
      cpu:
      disk:
      filesystem:
      memory:
      network:
      load:
      paging:
      processes:
      process:        # uncomment to enable per-process metrics (high cardinality)
```

The `process` scraper emits one set of metrics per running process. On a busy host this can produce thousands of data points per collection interval — enable it only if you specifically need per-process visibility and have sized your index accordingly.

For the full list of metrics each scraper emits and which are enabled by default, see the [hostmetricsreceiver documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver).

After any edit, restart the service:

```sh
sudo systemctl restart splunk-otel-collector
```

---

## Known Limitations, Constraints, and Troubleshooting

### Limitations

#### `file_log/etc` re-reads on every restart
`file_log/etc` has no `storage` set by design — it performs a one-time read of matching files under `/etc` on every collector start. This is intentional for capturing config state, but means it will re-send the same files after each restart. Do not add `storage` to this receiver; use `file_log/varlog` as a template if you need continuous tailing with checkpointing.

#### Journald requires group membership
The `journald` receiver will produce no data if the `splunk-otel-collector` service user is not a member of the `systemd-journal` group. The collector logs will show `journalctl command exited` with `exit status 1` from `journald/input.go`.

#### `confmap.enableMergeAppendOption` feature gate required for combined configs
When running both `agent_config.yaml` and a platform config together via `OTELCOL_OPTIONS`, the `confmap.enableMergeAppendOption` feature gate must be enabled. Without it, loading two `--config` files that define overlapping keys (e.g. `extensions`) will fail with a merge error. The installer sets this automatically.

#### HEC token default index
`--splunk-platform-logs-index` is optional. If omitted, `SPLUNK_PLATFORM_LOGS_INDEX` resolves to an empty string and the `index` field is omitted from the HEC payload — Splunk then routes events to the default index configured on the HEC token. If the token has no default index configured, events will be rejected silently. Either provide `--splunk-platform-logs-index` at install time, or ensure the HEC token has a default index set in Splunk.

---

### Verifying Ingestion

#### Verify logs are arriving in Splunk

Search for events with the sourcetypes set by the default receivers. The `linux:varlog` sourcetype covers `/var/log` files enabled by default:

```
index="<your-index>" sourcetype="linux:varlog"
```

To see all sourcetypes currently being ingested from this collector:

```
index="<your-index>" sourcetype="linux:*" | stats count by sourcetype
```

#### Verify metrics are arriving in Splunk

Metrics sent via `splunk_hec/metrics` are stored as metric events. Use `| mpreview` to inspect the metric index:

```
| mpreview index="<your-metrics-index>" | head 20
```

To check which metric names are being received:

```
| mpreview index="<your-metrics-index>" | stats count by metric_name
```

---

### Troubleshooting

For general log collection troubleshooting, see [Troubleshoot log collection](https://help.splunk.com/en/splunk-observability-cloud/manage-data/splunk-distribution-of-the-opentelemetry-collector/get-started-with-the-splunk-distribution-of-the-opentelemetry-collector/troubleshooting/troubleshoot-log-collection). The following covers issues specific to the platform log collection setup.

#### Collector fails to start: `mkdir /var: permission denied`
The `file_storage/filelogs` extension defaults to `/var/lib/otelcol/filelogs`. If `SPLUNK_FILE_STORAGE_EXTENSION_PATH` is not set and that directory does not exist or is not writable by the `splunk-otel-collector` user, startup fails. Fix:
```sh
sudo mkdir -p /var/lib/otelcol/filelogs
sudo chown splunk-otel-collector:splunk-otel-collector /var/lib/otelcol/filelogs
```
Or set `SPLUNK_FILE_STORAGE_EXTENSION_PATH` to a writable path in `/etc/otel/collector/splunk-otel-collector.conf`.

#### TLS error: `x509: cannot validate certificate for <IP> because it doesn't contain any IP SANs`

This error occurs when the HEC endpoint URL uses a raw IP address (e.g. `https://10.x.x.x:8088/services/collector`) but the server's TLS certificate was issued for a hostname, not an IP. The correct fix is to use the hostname in `--splunk-platform-url` instead of the IP address — in a normal Splunk deployment the certificate covers the server's DNS name.

If the server uses a self-signed or internal CA certificate, add the CA to the system trust store on the collector host so the certificate can be verified normally.

As a last resort for non-production environments, TLS verification can be disabled entirely:

```yaml
exporters:
  splunk_hec/logs:
    tls:
      insecure_skip_verify: true
```

This should not be used in production as it disables all certificate validation.

#### Exporting failed: `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`

This means the HEC endpoint accepted the connection but did not respond before the client timeout expired. It typically happens when Splunk is under heavy load and cannot process incoming data fast enough. The exporter will automatically retry — no data is lost as long as the queue does not fill up.

To give Splunk more time to respond, increase the HTTP client timeout in `splunk_logs_config_linux.yaml`:

```yaml
exporters:
  splunk_hec/logs:
    timeout: 30s
```

If retries are filling the queue and causing backpressure, also consider reducing the log volume or by disabling high-traffic receivers.

#### Checking collector logs

To follow collector logs in real time:
```sh
sudo journalctl -u splunk-otel-collector -f
```

To view the last 100 lines:
```sh
sudo journalctl -u splunk-otel-collector -n 100
```

#### No logs arriving in Splunk
1. Check the collector is running: `sudo systemctl status splunk-otel-collector`
2. Verify the HEC token and URL are correct in `/etc/otel/collector/splunk-otel-collector.conf`
3. Confirm the receiver is enabled in the pipeline — check `service.pipelines.logs/hec.receivers` in the config
4. Look for export errors in the collector logs (see above)
5. Common HEC errors: `401 Unauthorized` (bad token), `403 Forbidden` (token disabled or wrong index), `400 Bad Request` (malformed event — check `index` value)

#### Journald receiver produces no data
The collector log will show the following error when the service user lacks `systemd-journal` group membership:

```
Jun 02 09:41:29 ip-10-236-24-112 otelcol[103201]: 2026-06-02T09:41:29.462Z	error	journald/input.go:98	journalctl command exited	{"otelcol.component.id": "journald", "otelcol.component.kind": "receiver", "error": "exit status 1"}
```

Verify group membership:
```sh
groups splunk-otel-collector
```
If `systemd-journal` is not listed:
```sh
sudo usermod -aG systemd-journal splunk-otel-collector
sudo systemctl restart splunk-otel-collector
```

#### `Memory usage is above soft limit. Refusing data.`
This log message means the `memory_limiter` processor is rejecting incoming data because memory usage has exceeded the soft limit. The soft limit is derived from `SPLUNK_MEMORY_LIMIT_MIB` (90% of `SPLUNK_MEMORY_TOTAL_MIB`). To resolve:
- Increase `SPLUNK_MEMORY_TOTAL_MIB` to give the limiter more headroom
- Reduce the number of active receivers
- Lower `queue_size` on the `splunk_hec/logs` exporter

#### Metrics not arriving in Splunk Observability Cloud when using `--splunk-platform-url`

If you want to send both Splunk Platform logs **and** Splunk Observability Cloud metrics, you must provide an `access_token` (your o11y ingest token) to the installer. Without it, the installer skips the entire o11y configuration:

- `SPLUNK_ACCESS_TOKEN`, `SPLUNK_REALM`, `SPLUNK_INGEST_URL`, and related env vars are not written to the env file.
- `OTELCOL_OPTIONS` is set to load only the platform logs config — `agent_config.yaml` (which owns the o11y metrics pipeline) is never loaded.

**Correct usage** — provide both the o11y access token and the platform flags:

```sh
curl -sSL https://dl.signalfx.com/splunk-otel-collector.sh | sudo sh - \
  --realm us0 \
  --splunk-platform-token "YOUR_HEC_TOKEN" \
  --splunk-platform-url   "https://splunk.example.com:8088/services/collector" \
  -- YOUR_O11Y_ACCESS_TOKEN
```

The `--` before the access token is required when the token starts with a `-`. For tokens that do not start with `-`, the token can also be passed as a plain positional argument anywhere in the argument list.

If you installed without an access token and now want to add o11y metrics collection, re-run the installer with the access token included. The installer will rewrite the env file and update `OTELCOL_OPTIONS` to load both configs.

#### Config validation errors (`otelcol validate`, YAML issues, missing env vars)
For help diagnosing and fixing Collector configuration errors, see [Part 3: Troubleshoot common Collector configuration issues](https://help.splunk.com/en/splunk-observability-cloud/manage-data/splunk-distribution-of-the-opentelemetry-collector/get-started-with-the-splunk-distribution-of-the-opentelemetry-collector/collector-for-linux/configure-the-splunk-distribution-of-opentelemetry-collector-on-a-linux-host/part-3-troubleshoot-common-collector-configuration-issues).
