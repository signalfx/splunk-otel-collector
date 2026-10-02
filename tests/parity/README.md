# parity

A black-box parity test framework for the Splunk OTel Collector. Each case names
the event fields it cares about and has a checked-in **golden**: the events a
Splunk Universal Forwarder (the oracle) lands in Splunk for that input, reduced to
those fields. A normal run replays the collector (the candidate) against the
golden and checks it lands the same events. The golden is the single source of
truth for values; it is generated from UF out of band, not on every run.

## Model

The golden is derived from UF, so it is real Splunk output, not a hand-authored
guess. But UF is slow to install and boot, and re-running it every test would
make the oracle a moving target. So the framework splits into two steps:

1. **Generate** (`make update-goldens`, whenever a case is added or changed, or
   the pinned UF is updated): run UF against the case input and write its events
   to `golden.json`, reduced to the fields the case selects. This is the only step
   that needs a UF.
2. **Replay** (a normal run, and CI): run the collector against the same input
   and check the events it lands match `golden.json` on those same fields.

The two steps never run together: `-update` writes the golden and stops, so a
replay always runs against a golden that has been reviewed. Only one agent runs
per invocation, so both use the same index for a given case, which is what lets a
case select `index` like any other field.

Each run forwards into a Splunk instance started in Docker via testcontainers and
reads events back over the Splunk REST search API. Cases share the container and
the readback is scoped by index, so each case gets its own, named after its
directory (`parity_host`), and never sees the events of a case that ran before it.

Each case picks the transport for both sides, so they can differ on purpose. The
example case has UF send cooked S2S over HTTP (`conf/outputs.conf [httpout]`) to
`/services/collector/s2s` and the collector send JSON (`splunk_hec` exporter) to
`/services/collector/event`.

## Golden files

A golden holds only the fields the case's `expected` selects, one JSON object
per event:

```json
[
  {
    "raw": "initial text",
    "host": "myhost"
  }
]
```

Keeping the golden to those fields avoids checking in large, volatile payloads,
and it is what scopes the comparison: replay compares the candidate on exactly the
fields the golden holds and ignores everything else, so a sandbox path in `source`
or UF's auto-assigned `sourcetype` never enters a comparison unless the case
selects it.

Regenerate goldens after adding or changing a case, or bumping the oracle
version:

```sh
cd tests/parity
make update-goldens
```

The target downloads the UF version pinned in the `Makefile` into `.local/` and
runs `go test -update` against it, so the oracle is the same build everywhere
instead of whatever is installed on the machine. Bumping `SPLUNK_UF_VERSION`
means regenerating the goldens. To use an existing install instead, point
`PARITY_UF_DIR` at it and run `go test -update` directly.

Review the resulting `golden.json` diff before committing it. Regeneration is
deliberately not a CI job: it rewrites checked-in files that need review.

## Layout

```
tests/parity/
  Makefile             update-goldens: pinned UF install + `go test -update`
  parity.go            core types: Backend, Adapter, Validator, Record, HEC
  case.go              Case (test.yaml shape), token interpolation
  golden.go            LoadGolden / WriteGolden
  runner.go            RunCase / RunAgent: sandbox, hooks, capture loop
  validate.go          SubsetValidator
  replay_test.go       TestParity: replay (and -update to regenerate goldens)
  backend/splunk/       testcontainers-backed Splunk Backend
  adapter/uf/           UF adapter (the oracle, used only on -update)
  adapter/otelcol/      collector adapter (the candidate)
  tests/<case>/         one case: test.yaml + conf/ + golden.json, optional collector.yaml
```

## Interfaces

- **`Backend`** — the shared Splunk. `Start`/`Stop`, `HEC()` for the endpoint and
  token agents forward to, `Search(spl)` to read events back as `Record`s,
  `Clean(spl)` to delete. Implemented by `backend/splunk`.
- **`Adapter`** — drives one agent: `Prepare(configDir)` installs the case's
  interpolated config into the agent's own layout, `Start`/`Stop`/`Cleanup`
  manage lifecycle. `InstallDir()` reports the agent root, so a caller can check
  the agent is installed. Implemented by `adapter/uf` and `adapter/otelcol`.
- **`Validator`** — compares a reference capture against a candidate.
  `SubsetValidator` compares only the fields the reference sets, so a golden
  reduced by the case's filter scopes the comparison to those fields.
- **`Record`** — one indexed event, named as Splunk search surfaces it: `Raw`
  (`_raw`), `Host`, `Source`, `Sourcetype`, `Index`, `Time` (`_time`), and
  `Fields` for anything else. Its JSON tags shape the golden: a projected record
  marshals to only the keys the case selects.
- **`Case`** — one test loaded from `test.yaml`. **`AgentRun`** — one agent's
  participation: adapter, config files, target index, optional readback SPL.

## Writing a case

A case is a directory under `tests/`:

```
tests/host/
  test.yaml          metadata, shell hooks, the fields to compare
  conf/              the Splunk .conf structure, copied into the UF's etc/system/local
    inputs.conf
    outputs.conf
  collector.yaml     optional candidate config; without it the candidate reads conf/
  golden.json        generated
```

`conf/` is agent-agnostic Splunk config rather than anything UF-specific: every
`*.conf` in it is handed to the agent, so a parsing case adds
`props.conf`/`transforms.conf` with no framework change. It always configures the
oracle, and it configures the candidate too unless the case supplies a
`collector.yaml`. See [How the candidate is configured](#how-the-candidate-is-configured).

`test.yaml` holds everything that is not agent config:

```yaml
name: "Set host"
description: We use a custom host set when monitoring a file.
stage: alpha
setup: |                      # shell run before the agent starts
  echo "initial text" > foo.txt
script:                       # shell run after the agent starts
expected:                     # the fields to compare, and nothing else
  raw: true
  host: true
  # fields: [punct, "date_*"]  # search fields, by name or glob
os: ["darwin", "linux", "windows"]
```

`expected` is the case's filter: it names the fields the case is defined on, and
nothing else. `raw`, `host`, `source`, `sourcetype` and `index` are booleans;
`fields` selects `Record.Fields` keys by exact name or glob, so `date_*` takes
every `date_` field a search returns.

The same filter governs both steps: `make update-goldens` saves exactly these
fields of UF's events into `golden.json`, and a replay compares the candidate on
exactly these fields. The values live only in the golden, so there is nothing to
keep in sync. Leaving a field out keeps it out of both.

The `conf/` files, `collector.yaml`, `setup`, and `script` are interpolated
before use. Only the braced `${NAME}` form is a token, so shell expansions
(`$1`, `$(date)`), the collector's own `${env:NAME}` references, and any bare
mention of a name in event text pass through untouched:

| Token            | Value                                             |
| ---------------- | ------------------------------------------------- |
| `${BASE_DIR}`    | per-run sandbox working directory                 |
| `${CONFIG_DIR}`  | directory the run's config files are written to   |
| `${HEC_ENDPOINT}`| backend HEC endpoint, e.g. `https://127.0.0.1:...`|
| `${HEC_TOKEN}`   | backend HEC token                                 |
| `${INDEX}`       | the index this agent forwards to                  |

Splunk `.conf` files have no expansion of their own, which is why the framework
does the substitution rather than leaving it to each agent.

Then generate the golden with `make update-goldens` (see above) and commit it.

## How the candidate is configured

`collector.yaml` is optional. A case that has one is run from it. A case that
does not is run from the same `conf/` the oracle reads, through the
`splunk_inputs` receiver and `splunk_outputs` exporter, which covers the `.conf`
translation end to end: a stanza the collector maps differently shows up as a
mismatch instead of passing on a hand-written equivalent.

So a case asserts one of two things, and which one is visible from its directory:

- **with `collector.yaml`** — the golden is reachable with native collector
  config. Useful for a case whose point is the event shape rather than `.conf`
  handling.
- **without `collector.yaml`** — our `.conf` support reaches the golden from the
  same input UF got.

The collector config for a `conf/`-driven case is the framework's, not the
case's. Both components take a single `base_dir` and discover the stanzas
themselves, so there is nothing a case could vary, and keeping it out of the case
directory means a case cannot pin the translation it exists to test. The
framework materializes the case's `conf/` into `etc/system/local` under that
`base_dir` and enables `enableTARunner`, the alpha gate the two components are
registered behind.

`outputs.conf` is the one file such a case does not supply to the candidate: the
oracle ships events with `[httpout]`, a kind `splunk_outputs` skips, so the
framework renders a `[hecout]` stanza pointing at the backend instead.

## Running

Prerequisites for a normal (replay) run:

- Docker (the Splunk image is amd64-only; on Apple Silicon it runs under
  emulation).
- The collector binary. Build it with `make otelcol` from the repo root.

`make update-goldens` additionally downloads a UF, so it needs network access and
a platform the pinned tarball is published for (Linux amd64/arm64, macOS).

Environment variables:

| Variable              | Purpose                                                          |
| --------------------- | ---------------------------------------------------------------- |
| `PARITY_OTELCOL_BIN`  | collector binary. Defaults to `../../bin/otelcol`.               |
| `PARITY_UF_DIR`       | UF install root. Set by `make update-goldens`.                    |
| `PARITY_SPLUNK_IMAGE` | override the Splunk image (e.g. a locally cached tag).           |

Run:

```sh
cd tests/parity
go test -v -timeout 30m .
```

The first run pulls the Splunk image (~2.5GB) and boots it, so allow a few
minutes. CI replays on every PR via `.github/workflows/parity-test.yml`.

## Status

The suite replays the collector against UF-derived goldens on the fields each
case selects. The next milestone is adding cases and covering typed values, since
`Fields` values are strings today.
