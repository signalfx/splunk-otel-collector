# OBI Receiver (`obi`)

| Status                   |                            |
|--------------------------|----------------------------|
| Stability                | [alpha]                    |
| Supported pipeline types | traces, metrics            |
| Distributions            | [Splunk]                   |
| OS                       | Linux only (amd64, arm64)  |

The **OBI receiver** integrates [OpenTelemetry eBPF Instrumentation (OBI)](https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation)
into the Splunk OpenTelemetry Collector as a native receiver component. It uses Linux
eBPF probes to automatically collect distributed traces and RED metrics from running
processes **without code changes or application restarts**.

> **Alpha:** The OBI receiver is under active development. Configuration options,
> defaults, and behavior may change between releases.

## Requirements

### Operating System

OBI requires **Linux on amd64 or arm64**. The receiver is gated by a build tag
(`//go:build linux && (amd64 || arm64)`) — it is not compiled into the collector binary
on other platforms (Windows, macOS, ppc64le, etc.), so no `obi` receiver will be
registered on those targets.

**Minimum kernel version: 4.18**

### Linux Capabilities

OBI's eBPF probes require elevated Linux capabilities. The exact capabilities depend on
the kernel version and which observability features are used.

#### Core eBPF capabilities

| Kernel version | Minimum required capability(s)         |
|----------------|----------------------------------------|
| < 5.8          | `CAP_SYS_ADMIN`                        |
| 5.8 – 5.10     | `CAP_BPF`, `CAP_SYS_RESOURCE`          |
| ≥ 5.11         | `CAP_BPF`                              |

> `CAP_SYS_ADMIN` is a superset of all capabilities below. If granted, no other
> capabilities are needed regardless of kernel version.

#### Application Observability (App O11y) capabilities

Required when OBI instruments user-space application processes:

| Capability               | Purpose                                     |
|--------------------------|---------------------------------------------|
| `CAP_CHECKPOINT_RESTORE` | Process instrumentation                     |
| `CAP_DAC_READ_SEARCH`    | Reading process memory maps                 |
| `CAP_SYS_PTRACE`         | Attaching to running processes              |
| `CAP_PERFMON`            | eBPF perf event access                      |
| `CAP_NET_RAW`            | Raw socket access                           |
| `CAP_NET_ADMIN`          | Context propagation (when enabled)          |

#### Network Observability (Net O11y) capabilities

Required when OBI collects network-level flow data:

| Source type              | Required capabilities           |
|--------------------------|---------------------------------|
| TC (traffic control)     | `CAP_PERFMON`, `CAP_NET_ADMIN`  |
| Socket                   | `CAP_NET_RAW`                   |

#### Granting capabilities

**Bare metal / VMs** — grant capabilities on the collector binary with `setcap`:

```bash
# Kernel >= 5.8 (App O11y + context propagation)
sudo setcap cap_bpf,cap_checkpoint_restore,cap_dac_read_search,\
cap_sys_ptrace,cap_perfmon,cap_net_raw,cap_net_admin+ep \
/usr/bin/otelcol

# Kernel < 5.8 (all features via CAP_SYS_ADMIN)
sudo setcap cap_sys_admin+ep /usr/bin/otelcol
```

**Docker** — pass `--cap-add` flags:

```bash
docker run \
  --cap-add BPF \
  --cap-add CHECKPOINT_RESTORE \
  --cap-add DAC_READ_SEARCH \
  --cap-add SYS_PTRACE \
  --cap-add PERFMON \
  --cap-add NET_RAW \
  --cap-add NET_ADMIN \
  splunk/otelcol:latest
```

**Kubernetes** — add capabilities in the DaemonSet pod `securityContext`:

```yaml
containers:
  - name: otel-collector
    securityContext:
      capabilities:
        add:
          - BPF
          - CHECKPOINT_RESTORE
          - DAC_READ_SEARCH
          - SYS_PTRACE
          - PERFMON
          - NET_RAW
          - NET_ADMIN
```

#### `enforce_sys_caps`

By default (`enforce_sys_caps: false`), OBI logs a warning if required capabilities
are missing but continues running. Features that require missing capabilities will
silently produce no data.

Set `enforce_sys_caps: true` to make the receiver exit immediately on insufficient
capabilities:

```yaml
receivers:
  obi:
    enforce_sys_caps: true
```

## Configuration

### Minimal configuration

The simplest configuration instruments all processes listening on a given port:

```yaml
receivers:
  obi:
    open_port: "8080"
```

An empty `obi: {}` section starts the receiver but instruments nothing until
`discovery` rules match processes.

### Full pipeline example

```yaml
receivers:
  obi:
    open_port: "8080,9090"   # comma-separated ports or ranges
    service_name: my-service
    service_namespace: production
    enforce_sys_caps: false  # log warning if caps missing (default)
    log_level: INFO          # DEBUG | INFO | WARN | ERROR

exporters:
  otlp:
    endpoint: "https://ingest.us0.signalfx.com:443"
    headers:
      X-SF-Token: "${SPLUNK_ACCESS_TOKEN}"

service:
  pipelines:
    traces:
      receivers: [obi]
      exporters: [otlp]
    metrics:
      receivers: [obi]
      exporters: [otlp]
```

### Service discovery

Instead of a fixed port, use `discovery` to automatically find and instrument
matching processes as they start and stop:

```yaml
receivers:
  obi:
    discovery:
      poll_interval: 30s
      instrument:
        - exe_path: "/usr/bin/python*"
        - exe_path: "/usr/local/bin/node"
          open_ports: "3000"
        - name: "my-java-app"
          languages: [java]
      exclude_otel_instrumented_services: true  # skip services already sending OTLP
```

#### Process selectors (`instrument` / `exclude_instrument`)

Each entry is a [`GlobAttributes`](https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/v0.6.0/pkg/appolly/services/attr_glob.go) object.
OBI instruments a process if it matches **all** non-empty fields in **any** single entry
(logical OR across entries; AND within an entry).

| Field                 | Type        | Description                                               |
|-----------------------|-------------|-----------------------------------------------------------|
| `exe_path`            | glob        | Path to the process executable                            |
| `name`                | glob        | Process name                                              |
| `namespace`           | glob        | Process namespace                                         |
| `open_ports`          | port list   | Ports the process is listening on (`"8080"`, `"8000-8099"`) |
| `languages`           | string list | `go`, `java`, `python`, `ruby`, `nodejs`, `dotnet`, `rust` |
| `target_pids`         | int list    | Target specific process IDs                               |
| `cmd_args`            | glob        | Match against process command-line arguments              |
| `k8s_pod_labels`      | map         | Kubernetes pod label key/value selectors                  |
| `k8s_pod_annotations` | map         | Kubernetes pod annotation key/value selectors             |
| `containers_only`     | bool        | Only match containerized processes                        |

**Example — instrument specific Kubernetes pods by label:**

```yaml
discovery:
  instrument:
    - k8s_pod_labels:
        app: "checkout"
      namespace: "production"
    - k8s_pod_labels:
        app: "payment"
      containers_only: true
```

### Configuration reference

| Field                                          | Default       | Description                                                 |
|------------------------------------------------|---------------|-------------------------------------------------------------|
| `open_port`                                    | —             | Port(s) to instrument (comma-separated or ranges)           |
| `service_name`                                 | auto-detected | Override the instrumented service name                      |
| `service_namespace`                            | auto-detected | Override the instrumented service namespace                 |
| `log_level`                                    | `INFO`        | OBI log verbosity: `DEBUG`, `INFO`, `WARN`, `ERROR`         |
| `enforce_sys_caps`                             | `false`       | Exit if required Linux capabilities are missing             |
| `discovery.poll_interval`                      | —             | Re-scan interval for new processes                          |
| `discovery.instrument`                         | —             | Process selectors to instrument (see above)                 |
| `discovery.exclude_instrument`                 | —             | Process selectors to exclude from instrumentation           |
| `discovery.exclude_otel_instrumented_services` | `false`       | Skip processes already emitting OTLP data                   |

For advanced options (OTLP export endpoints, Prometheus export, Kubernetes metadata
decoration, Node.js/JVM agent settings, etc.) refer to the
[OBI upstream documentation](https://opentelemetry.io/docs/zero-code/) and the
[`Config` struct in `pkg/obi/config.go`](https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/v0.6.0/pkg/obi/config.go).
