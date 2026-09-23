# parity

A black-box parity test framework for the Splunk OTel Collector. Each case has a
checked-in **golden**: the events a Splunk Universal Forwarder (the oracle) lands
in Splunk for that input, reduced to the fields the case asserts. A normal run
replays the collector (the candidate) against the golden and checks it lands the
same events. The golden is the single source of truth; it is generated from UF
out of band, not on every run.

## Model

The golden is derived from UF, so it is real Splunk output, not a hand-authored
guess. But UF is slow to install and boot, and re-running it every test would
make the oracle a moving target. So the framework splits into two steps:

1. **Generate** (`go test -update`, done occasionally by a maintainer): run UF
   against the case input, project its events down to the asserted fields, and
   write `golden.json`. This is the only step that needs a UF install.
2. **Replay** (a normal run, and CI): run the collector against the same input
   and assert the events it lands match `golden.json` on the asserted fields.

Both agents forward into a single Splunk instance started in Docker via
testcontainers, each to its own index (`parity_uf`, `parity_uc`), and validation
reads events back over the Splunk REST search API. The two agents use different
native transports and config formats on purpose:

- UF sends cooked S2S over HTTP (`outputs.conf [httpout]`) to
  `/services/collector/s2s`.
- The collector sends JSON (`splunk_hec` exporter) to `/services/collector/event`.

## Golden files

A golden holds only the fields the case asserts, one JSON object per event:

```json
[
  {
    "raw": "initial text",
    "host": "myhost"
  }
]
```

Keeping the golden to the asserted fields is deliberate: it avoids checking in
large, volatile payloads, and it doubles as the comparison filter. Replay
compares the candidate on exactly the fields the golden sets and ignores
everything else, so the differing index and UF's auto-assigned
`source`/`sourcetype` never enter the comparison unless a case asserts them.

Regenerate goldens after changing a case input or the oracle version:

```sh
cd tests/parity
PARITY_UF_DIR=/path/to/splunkforwarder go test -update -v -timeout 30m .
```

Review the resulting `golden.json` diff before committing it.

## Layout

```
tests/parity/
  parity.go            core types: Backend, Adapter, Validator, Record, HEC
  case.go              Case (test.yaml shape), token interpolation
  golden.go            Project / LoadGolden / WriteGolden
  runner.go            RunAgent: sandbox, hooks, capture loop
  validate.go          SubsetValidator
  parity_test.go       TestParity: replay (and -update to regenerate goldens)
  backend/splunk/       testcontainers-backed Splunk Backend
  adapter/uf/           UF adapter (the oracle, used only on -update)
  adapter/otelcol/      collector adapter (the candidate)
  tests/<case>/         one case: test.yaml + collector.yaml + golden.json
```

## Interfaces

- **`Backend`** — the shared Splunk. `Start`/`Stop`, `HEC()` for the endpoint and
  token agents forward to, `Search(spl)` to read events back as `Record`s,
  `Clean(spl)` to delete. Implemented by `backend/splunk`.
- **`Adapter`** — drives one agent: `Prepare(configDir)` installs the case's
  interpolated config into the agent's own layout, `Start`/`Stop`/`Cleanup`
  manage lifecycle. `InstallDir()` supplies the `AGENT_DIR` token. Implemented by
  `adapter/uf` and `adapter/otelcol`.
- **`Validator`** — compares a reference capture against a candidate.
  `SubsetValidator` asserts only the fields the reference sets; empty fields and
  volatile keys (indextime, stream/ACK ids) are ignored.
- **`Record`** — one indexed event, named as Splunk search surfaces it: `Raw`
  (`_raw`), `Host`, `Source`, `Sourcetype`, `Index`, `Time` (`_time`), and
  `Fields` for anything else. Its JSON tags shape the golden: a projected record
  marshals to only the asserted keys.
- **`Case`** — one test loaded from `test.yaml`. **`AgentRun`** — one agent's
  participation: adapter, config files, target index, optional readback SPL.

## Writing a case

Add a directory under `tests/` with `test.yaml`, `collector.yaml`, and a
generated `golden.json`. `test.yaml`:

```yaml
name: "Set host"
description: We use a custom host set when monitoring a file.
stage: alpha
conf: |                       # UF inputs.conf fragment
  [monitor:///BASE_DIR/foo.txt]
  host=myhost
  index=INDEX
setup: |                      # shell run before the agent starts
  echo "initial text" > foo.txt
script:                       # shell run after the agent starts
assert: [raw, host]           # Record fields to compare against the golden
os: ["darwin", "linux", "windows"]
```

`assert` names the fields to compare: `raw`, `host`, `source`, `sourcetype`,
`index`, or `field:<key>` for a search field. Those are the fields `-update`
saves from UF and replay compares. `collector.yaml` in the same directory is the
candidate's config that should produce the same events.

`conf`, `collector.yaml`, `setup`, and `script` are interpolated with these
tokens before use:

| Token          | Value                                             |
| -------------- | ------------------------------------------------- |
| `BASE_DIR`     | per-run sandbox working directory                 |
| `AGENT_DIR`    | agent install root                                |
| `HEC_ENDPOINT` | backend HEC endpoint, e.g. `https://127.0.0.1:...`|
| `HEC_TOKEN`    | backend HEC token                                 |
| `INDEX`        | the index this agent forwards to                  |

Then generate the golden with `-update` (see above) and commit it.

## Running

Prerequisites for a normal (replay) run:

- Docker (the Splunk image is amd64-only; on Apple Silicon it runs under
  emulation).
- The collector binary. Build it with `make otelcol` from the repo root.

Regenerating goldens (`-update`) additionally needs a UF install via
`PARITY_UF_DIR`.

Environment variables:

| Variable              | Purpose                                                        |
| --------------------- | -------------------------------------------------------------- |
| `PARITY_OTELCOL_BIN`  | collector binary. Defaults to `../../bin/otelcol`.             |
| `PARITY_UF_DIR`       | UF install root. Only needed for `-update`.                    |
| `PARITY_SPLUNK_IMAGE` | override the Splunk image (e.g. a locally cached tag).         |

Run:

```sh
cd tests/parity
go test -v -timeout 30m .
```

The first run pulls the Splunk image (~2.5GB) and boots it, so allow a few
minutes. CI replays on every PR via `.github/workflows/parity-test.yml`;
regenerating goldens is a manual `workflow_dispatch`.

## Status

The suite replays the collector against UF-derived goldens on the asserted
fields. The next milestone is porting more of the 1spl corpus and widening the
asserted fields (index-time fields, structured `Fields` values).
