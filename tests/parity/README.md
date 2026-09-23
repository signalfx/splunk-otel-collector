# parity

A black-box parity test framework for UF-compatible agents. It runs the same
inputs through a Splunk Universal Forwarder (the oracle) and a candidate agent
(the Splunk OTel Collector), forwards both into one real Splunk instance, and
compares the events that land there.

## Model

Both agents forward into a single Splunk instance started in Docker via
testcontainers. Each agent writes to its own index (`parity_uf`, `parity_uc`),
so a run needs no clean-between-agents step. Validation reads events back over
the Splunk REST search API and diffs them.

The two agents use different native transports and config formats on purpose:

- UF sends cooked S2S over HTTP (`outputs.conf [httpout]`) to
  `/services/collector/s2s`.
- The collector sends JSON (`splunk_hec` exporter) to `/services/collector/event`.

Same inputs, same resulting Splunk events. Agents are never a matrix axis: a run
is always oracle vs candidate.

## Layout

```
tests/parity/
  parity.go            core types: Backend, Adapter, Validator, Record, HEC
  case.go              Case (test.yaml shape), Expected, token interpolation
  runner.go            RunAgent / RunCase: sandbox, hooks, capture loop
  validate.go          SubsetValidator
  normalize.go         Normalizer, NormalizingValidator for direct parity
  parity_test.go       TestParity: loads tests/*/test.yaml and runs both agents
  backend/splunk/       testcontainers-backed Splunk Backend
  adapter/uf/           UF adapter (the oracle)
  adapter/otelcol/      collector adapter (the candidate)
  tests/<case>/         one directory per case (test.yaml + collector.yaml)
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
  volatile keys (indextime, stream/ACK ids) are ignored. `NormalizingValidator`
  wraps it with a `Normalizer` that blanks fields the oracle and candidate assign
  differently by construction (index, source, sourcetype), so a live UF capture
  can be the reference in a direct parity check.
- **`Record`** — one indexed event, named as Splunk search surfaces it: `Raw`
  (`_raw`), `Host`, `Source`, `Sourcetype`, `Index`, `Time` (`_time`), and
  `Fields` for anything else.
- **`Case`** — one test loaded from `test.yaml`. **`AgentRun`** — one agent's
  participation: adapter, config files, target index, optional readback SPL.

## Writing a case

Add a directory under `tests/`. `test.yaml` is the ported 1spl case shape:

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
expected:                     # reference event, in Splunk terms
  raw: "initial text"
  host: myhost
os: ["darwin", "linux", "windows"]
```

`collector.yaml` in the same directory is the candidate's config that should
produce the same event.

Both files, plus `conf`, `setup`, and `script`, are interpolated with these
tokens before use:

| Token          | Value                                             |
| -------------- | ------------------------------------------------- |
| `BASE_DIR`     | per-run sandbox working directory                 |
| `AGENT_DIR`    | agent install root                                |
| `HEC_ENDPOINT` | backend HEC endpoint, e.g. `https://127.0.0.1:...`|
| `HEC_TOKEN`    | backend HEC token                                 |
| `INDEX`        | the index this agent forwards to                  |

The `expected` block is authored in Splunk-event terms (the same fields a search
returns). Only the fields you set are asserted.

## Running

Prerequisites:

- Docker (the Splunk image is amd64-only; on Apple Silicon it runs under
  emulation).
- A Splunk Universal Forwarder install, located via `PARITY_UF_DIR`.
- The collector binary. Build it with `make otelcol` from the repo root.

Environment variables:

| Variable              | Purpose                                                        |
| --------------------- | -------------------------------------------------------------- |
| `PARITY_UF_DIR`       | UF install root. If unset, the test is skipped.                |
| `PARITY_OTELCOL_BIN`  | collector binary. Defaults to `../../bin/otelcol`.             |
| `PARITY_SPLUNK_IMAGE` | override the Splunk image (e.g. a locally cached tag).         |

Run:

```sh
cd tests/parity
PARITY_UF_DIR=/path/to/splunkforwarder go test -v -timeout 30m .
```

The first run pulls the Splunk image (~2.5GB) and boots it, so allow a few
minutes. CI runs this via `.github/workflows/parity-test.yml`.

## Status

The suite now compares the candidate directly against the UF oracle: the UF
capture is checked against the case's authored `expected` (so a broken oracle is
distinguishable from a parity gap), then the candidate is compared field by field
against what UF actually landed. `ParityNormalizer` blanks the `index`, `source`,
and `sourcetype` the two agents assign differently by construction, leaving `raw`
and `host` (plus any asserted custom fields) as the parity signal.

Next: map the collector's `com.splunk.source` / `com.splunk.sourcetype` so those
become real parity assertions instead of normalized-away, populate `Record.Fields`
for index-time fields, and port the full 1spl suite.
