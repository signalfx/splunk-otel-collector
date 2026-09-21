# OTel Collector vs Universal Forwarder — TA Correctness Testing

## Goal

Verify that the Splunk OTel Collector running Technical Addons via tarunner produces
equivalent data in Splunk compared to the Universal Forwarder running the same TAs.
Both agents run from identical TA configuration files. A pytest-based test suite queries
both Splunk indexes and asserts parity across sourcetypes, event counts, fields, and timestamps.

---

## Architecture

```
VM A (Linux)                              VM B (Linux)
┌─────────────────────────────┐           ┌──────────────────────────────────────┐
│ Splunk Universal Forwarder  │           │ splunk-otel-collector                │
│                             │           │ (current repository checkout)        │
│ $SPLUNK_HOME/etc/apps/      │           │                                      │
│   <TA>/                     │           │                                      │
│     default/inputs.conf     │           │ $SPLUNK_HOME/etc/apps/               │
│     local/inputs.conf  ─────┼──same──▶  │   <TA>/                              │
│   system/local/outputs.conf │           │     default/inputs.conf              │
│   (S2S → VM C:9997)         │           │     local/inputs.conf                │
└─────────────────────────────┘           │   system/local/outputs.conf          │
         S2S                              │   (HEC → VM C:8088)                  │
          │                               │                                      │
          │                               │ splunk_inputs receiver reads TA conf │
          │                               │ natively via tarunner                │
          │                               │ splunk_outputs exporter reads        │
          │                               │ outputs.conf → HEC exporter          │
          │                               └──────────────────────────────────────┘
          │                                          HEC
          ▼                                           │
    index=uf_<ta>                              index=otel_<ta>
          │                                           │
          └─────────────────┬─────────────────────────┘
                            ▼
             ┌──────────────────────────────────┐
             │  VM C: Splunk Enterprise         │
             │  + TA installed (search-time     │
             │    props.conf field extraction)  │
             └──────────────────────────────────┘
                            │
                     pytest test suite
                     (runs from local machine)
```

---

## Components

### splunk-otel-collector
- Provides the `splunkinputsreceiver` and `splunkoutputsexporter` components
- Both gated behind feature gate `enableTARunner` (StageAlpha, registered from v0.158.0)
- The test framework builds the current collector checkout directly
- Binary started with `--feature-gates=+enableTARunner`

### TARunner implementation in the collector
- Reads `inputs.conf`, `props.conf`, `transforms.conf` using Splunk conf precedence:
  `system/default` < `ta/default` < `ta/local` < `system/local`
- Discovers all `splunk_ta_*` directories under `$SPLUNK_HOME/etc/apps/` automatically
- Watches for TA add/remove/change via fsnotify (500ms debounce), restarts sub-receivers
- Receivers: `monitorreceiver` (`monitor://`), `scriptreceiver` (`script://`),
  `batchreceiver`, `tcpreceiver`, `udpreceiver`, `wineventlogreceiver`
- `splunkoutputsexporter` reads `outputs.conf` from `$SPLUNK_HOME` and creates a HEC exporter

---

## Steps

### 1. Build

Build the `otelcol` binary from the collector checkout. The build script does
not switch branches or modify `go.mod`/`go.sum`.

```bash
make build
# produces: bin/otelcol_linux_amd64
```

Place the binary under `bin/otelcol_linux_amd64` before running `make setup-otel`.

Script: `scripts/build_otel.sh`

### 2. Install Splunk Enterprise (VM C)

- Downloads `splunk-<VERSION>-<BUILD_HASH>-linux-amd64.tgz` from download.splunk.com
- Sets admin password via `user-seed.conf` (more reliable than `--seed-passwd` on 9.x)
- Creates indexes `uf_<ta>` and `otel_<ta>` via REST API
- Enables HEC on port 8088 with a known token (SSL disabled)
- S2S receiving on port 9997 is enabled by default — no configuration needed
- Installs the TA on the search head for search-time field extraction (`props.conf`)
- Supports TA as extracted directory or `.tgz`/`.tar.gz` tarball under `tas/`

```bash
make setup-splunk
```

Script: `scripts/install_splunk.sh`

### 3. Install Universal Forwarder (VM A)

- Downloads `splunkforwarder-<VERSION>-<BUILD_HASH>-linux-amd64.tgz` from download.splunk.com
- Writes `system/local/outputs.conf` pointing to VM C port 9997
- Deploys TA from `tas/<TA>/` or `tas/<TA>.tgz`
- Appends `[default]\nindex = <UF_INDEX>` to TA `local/inputs.conf` if not already set
- Stops UF before `enable boot-start`, then restarts

```bash
make setup-uf
```

Script: `scripts/install_uf.sh`

### 4. Install OTel Collector (VM B)

- Copies `bin/otelcol_linux_amd64` to VM B via scp
- Creates `$SPLUNK_HOME` layout with correct ownership for the SSH user
- Deploys TA from `tas/<TA>/` or `tas/<TA>.tgz`
- Appends `[default]\nindex = <OTEL_INDEX>` to TA `local/inputs.conf` if not already set
- Writes `system/local/outputs.conf` with `[httpout]` stanza → HEC on VM C port 8088
- Writes `$SPLUNK_HOME/etc/otel_collector.yaml` (inline, no separate config file needed)
- Installs and starts systemd service: `otelcol --config ... --feature-gates=+enableTARunner`
- Installs TA runtime dependencies: `net-tools`, `lsof`, `sysstat`, `auditd`, `ntpsec-ntpdate`/`ntpdate`, `lastlog2`

```bash
make setup-otel
```

Script: `scripts/install_otel.sh`

### 5. TA Management

Deploy a TA to both VM A and VM B in sync. Restarts both agents after deploy.

```bash
make add-ta TA=Splunk_TA_nix
```

Script: `scripts/add_ta.sh`

- Resolves TA source: extracted directory or `.tgz`/`.tar.gz` tarball
- Copies to both VMs, sets index in `local/inputs.conf` on each
- Restarts UF on VM A and otelcol systemd service on VM B

### 5a. Deterministic end-to-end input

The repository includes `tas/Splunk_TA_otel_uf_test`, a small test TA with one
enabled stanza in `default/inputs.conf`:

```text
[monitor:///home/splunker/otel_uf_test.log]
sourcetype = otel_uf_test
index = __INDEX__
```

During deployment, `__INDEX__` is replaced independently on each agent:

```text
VM A / UF:  uf_index
VM B / OTel: otel_index
```

The input file is intentionally outside `/var/log`, so the test does not
depend on distro-specific log permissions, blacklists, or TA scripted inputs.
Append the same event to both machines with:

```bash
make append-test-event \
  EVENT='otel_uf_test_id=case-001 user=alice src=10.0.0.1 dest=10.0.0.2 action=allowed app=ssh'
```

The corresponding test `tests/tas/test_splunk_ta_otel_uf_test.py` requires a
non-zero exact count match, then checks field and timestamp parity. This makes
an un-ingested event fail instead of passing as `0 == 0`.

### 6. Run Tests

```bash
pip install -r requirements.txt
make compare
```

Runs `pytest tests/` against the live Splunk instance on VM C.

---

## Test Framework

**Stack:** pytest + splunk-sdk

### Test axes

| Test              | SPL                                            | Pass condition                                            |
|-------------------|------------------------------------------------|-----------------------------------------------------------|
| Sourcetype parity | `\| stats count by sourcetype` on both indexes | No sourcetype present in UF index missing from OTel index |
| Event count       | `\| stats count` per sourcetype                | Within 5%                                                 |
| Field presence    | `\| fieldsummary` per sourcetype               | No field present in UF missing from OTel (>50% coverage)  |
| Timestamp delta   | `avg(eval(_indextime-_time))` per sourcetype   | avg offset ≤ 60s                                          |

### Structure

```
tests/
├── conftest.py               # fixtures: Splunk client, uf_index, otel_index
├── lib/
│   ├── __init__.py
│   ├── splunk.py             # thin splunk-sdk search wrapper
│   └── assertions.py        # sourcetype_parity, count_parity, fields_parity, timestamp_delta
└── tas/
    ├── __init__.py
    └── test_splunk_ta_nix.py
```

---

## Directory Layout

```
otel-uf-tests/
├── PLAN.md
├── config.env.example
├── config.env                        # gitignored
├── .gitignore
├── Makefile
├── requirements.txt
├── bin/                              # gitignored — place otelcol_linux_amd64 here
├── scripts/
│   ├── build_otel.sh                 # build otelcol from the parent repository
│   ├── install_splunk.sh             # VM C: Splunk + indexes + HEC + TA
│   ├── install_uf.sh                 # VM A: UF + TA
│   ├── install_otel.sh               # VM B: otelcol binary + TA + systemd service
│   └── add_ta.sh                     # deploy TA to both VMs in sync
├── tas/
│   └── Splunk_TA_nix/
│       └── local/
│           └── inputs.conf           # which inputs to enable for testing
│   └── Splunk_TA_otel_uf_test/
│       └── default/
│           ├── inputs.conf           # one deterministic file monitor
│           └── props.conf            # key/value search-time extraction
│   └── Splunk_TA_nix.tgz             # gitignored — place TA tarball here instead
└── tests/
    ├── conftest.py
    ├── lib/
    │   ├── __init__.py
    │   ├── splunk.py
    │   └── assertions.py
    └── tas/
        ├── __init__.py
        └── test_splunk_ta_nix.py
        └── test_splunk_ta_otel_uf_test.py
```

---

## TA Rollout Order

| Phase | TA                  | Inputs                                                  | Notes                                                             |
|-------|---------------------|---------------------------------------------------------|-------------------------------------------------------------------|
| 1     | `Splunk_TA_nix`     | `monitor:///var/log`, `monitor:///home/*/.bash_history` | File monitor only                                                 |
| 2     | `Splunk_TA_nix`     | Scripted inputs (`cpu`, `ps`, `df`, `vmstat`, ...)      | After phase 1 passes                                              |
| 3     | `Splunk_TA_windows` | Windows Event Log, perfcounters, file monitors          | Separate Windows VMs, existing `otel-config.yaml` from prior work |

---

## Known Issues / Workarounds

| Issue                                                              | Workaround                                                   |
|--------------------------------------------------------------------|--------------------------------------------------------------|
| Splunk 9.4.x download URL includes build hash                      | Set `SPLUNK_BUILD_HASH=6b4ebe426ca6` in `config.env`         |
| `--seed-passwd` silently rejected on Splunk 9.x (complexity rules) | Use `user-seed.conf` written before first start              |
| `go generate` fails during build (mdatagen not installed)          | `build_otel.sh` calls `go build` directly, skipping generate |
| `find -mindepth` not available on macOS BSD find                   | TA extraction uses portable glob `for d in "$TMPDIR"/*/`     |
| S2S port 9997 enabled by default                                   | No REST API call needed to enable it                         |

---

## Out of Scope

- Performance/throughput testing — tooling already exists separately
- filelog OTel mode — tarunner mode only
- Golden file snapshots — follow-up once baseline is stable
- HF mode (cooked data / index-time field extraction by tarunner) — experimental, tracked separately
