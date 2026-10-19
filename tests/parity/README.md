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

Each run forwards into a real Splunk and reads events back over the Splunk REST
search API. By default the suite starts one in Docker via testcontainers;
setting `PARITY_SPLUNK_HEC_URL` and `PARITY_SPLUNK_MGMT_URL` points it at an
existing instance instead and skips the container. Cases share the one backend
and the readback is scoped by index, so each case gets its own, named after its
directory (`parity_host`), and never sees the events of a case that ran before it.

An external Splunk must be **empty and privileged**. The suite creates the
per-case indexes at startup and searches them over all time, and three goldens
pin the index name they expect (`tests/monitor-index`, `tests/tcp-event`,
`tests/udp-event`), so the names cannot be varied per run. Events left in
`parity_*` by an earlier run are counted as the candidate's own and fail the
comparison with a `field "count"` mismatch that gives no hint why. Use a fresh
instance per run.

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
    "host": "myhost",
    "source": "${BASE_DIR}/foo.txt"
  }
]
```

Keeping the golden to those fields avoids checking in large, volatile payloads,
and it is what scopes the comparison: replay compares the candidate on exactly the
fields the golden holds and ignores everything else, so UF's auto-assigned
`sourcetype` never enters a comparison unless the case selects it.

`source` is held as `${BASE_DIR}/...`: each agent runs in its own sandbox, so the
framework rewrites the sandbox prefix of a captured path back to the token both
when it writes a golden and when it reads a candidate back.

A golden is written in the order a capture is sorted into, by event text and
then source, not in the order the search returned. Event order is therefore not
asserted: a candidate stamps one time on a batch, so the index is free to return
those events either way round.

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
  backend/splunk/       Splunk Backend: a container it starts, or an existing instance
  adapter/uf/           UF adapter (the oracle, used only on -update)
  adapter/otelcol/      collector adapter (the candidate)
  tests/<case>/         one case: test.yaml + conf/ + golden.json, optional collector.yaml
```

## Interfaces

- **`Backend`** — the shared Splunk. `Start`/`Stop`, `HEC()` for the endpoint and
  token agents forward to, `Search(spl)` to read events back as `Record`s,
  `Clean(spl)` to delete. Implemented by `backend/splunk`, which serves both a
  Splunk it starts in Docker and one supplied by the environment; on the latter
  `Start` only resolves the endpoints and creates the indexes, and `Stop` is a
  no-op.
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
  conf/              the Splunk .conf structure, installed into the agent's etc/
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

A top-level `*.conf` is system config; a nested path is mirrored, so
`conf/apps/my_app/local/inputs.conf` installs as an app rather than as system
config. `tests/app-scoped-input` is an example.

`test.yaml` holds everything that is not agent config:

```yaml
name: "Set host"
description: We use a custom host set when monitoring a file.
setup: |                      # shell run before the agent starts
  echo "initial text" > foo.txt
script:                       # shell run after the agent starts
validate:                     # shell that must exit 0 once the capture settles
expected:                     # the fields to compare, and nothing else
  raw: true
  host: true
  # fields: [punct, "date_*"]  # search fields, by name or glob
os: ["linux"]                 # platforms to run on; unset runs on every platform
```

`os` is a filter for a platform-specific case: a case that leaves it out runs
wherever the suite is run, and a case that sets it is skipped on a platform it
does not name.

`expected` is the case's filter: it names the fields the case is defined on, and
nothing else. `raw`, `host`, `source`, `sourcetype` and `index` are booleans;
`fields` selects `Record.Fields` keys by exact name or glob, so `date_*` takes
every `date_` field a search returns.

The same filter governs both steps: `make update-goldens` saves exactly these
fields of UF's events into `golden.json`, and a replay compares the candidate on
exactly these fields. The values live only in the golden, so there is nothing to
keep in sync. Leaving a field out keeps it out of both.

### Asserting what is not an event

Some behavior leaves no event. A `batch://` stanza consumes its input, and the
golden is an array of events, so it cannot carry a deletion. `validate` is shell
that must exit 0:

```yaml
validate: |
  test ! -e foo.txt
```

It runs for whichever agent is running, so `-update` checks it against UF and a
replay against the candidate. A hook that does not hold for UF is a broken case
and fails at generation time, which is what keeps it an assertion about both
agents rather than only the candidate.

Two things to know when writing one:

- **It runs with the agent still going**, after the event capture settles, so it
  asserts what the agent did while reading rather than what it cleaned up on
  shutdown.
- **It is retried until it holds**, because an effect can land after the events
  that preceded it. Write a hook that tests current state and let the runner
  retry; do not hand-roll a loop.

The `conf/` files, `collector.yaml`, `setup`, and `script` are interpolated
before use. Only the braced `${NAME}` form is a token, so shell expansions
(`$1`, `$(date)`), the collector's own `${env:NAME}` references, and any bare
mention of a name in event text pass through untouched:

| Token            | Value                                             |
| ---------------- | ------------------------------------------------- |
| `${BASE_DIR}`    | per-run sandbox working directory                 |
| `${HEC_ENDPOINT}`| backend HEC endpoint, e.g. `https://127.0.0.1:...`|
| `${HEC_TOKEN}`   | backend HEC token                                 |
| `${INDEX}`       | the index this agent forwards to                  |

Splunk `.conf` files have no expansion of their own, which is why the framework
does the substitution rather than leaving it to each agent.

Then generate the golden with `make update-goldens` (see above) and commit it.

## How the candidate is configured

`collector.yaml` is optional. A case that has one is run from it. A case that
does not is run from the same `conf/` the oracle reads, through the `splunkconf`
config provider, which covers the `.conf` translation end to end: a stanza the
provider maps differently shows up as a mismatch instead of passing on a
hand-written equivalent.

So a case asserts one of two things, and which one is visible from its directory:

- **with `collector.yaml`** — the golden is reachable with native collector
  config. Useful for a case whose point is the event shape rather than `.conf`
  handling.
- **without `collector.yaml`** — our `.conf` support reaches the golden from the
  same input UF got.

Such a case has no collector config at all. Its only `--config` is
`splunkconf://${SPLUNK_HOME}/etc`, and the provider emits the receivers,
exporters and pipeline from the stanzas it finds, so there is nothing a case
could vary and nothing in the case directory that could pin the translation it
exists to test. Handed a case's `*.conf`, the adapter installs them into
`etc/system/local` of a sandbox `$SPLUNK_HOME`, the same way the UF adapter
installs them into its own install, and resolves `${SPLUNK_HOME}` in the URI to
that tree, standing in for the shell that would expand it in a real deployment.
The case also enables `enableTARunner`, the alpha gate the emitted `splunk_*`
receivers are registered behind.

`outputs.conf` is the one file such a case does not supply to the candidate: the
oracle ships events with `[httpout]`, which the provider does not map, so the
framework renders a `[hecout]` stanza pointing at the backend instead.

## Running

Prerequisites for a normal (replay) run:

- A Splunk. Either Docker, which lets the suite start its own (the image is
  amd64-only; on Apple Silicon it runs under emulation), or an existing instance
  selected with `PARITY_SPLUNK_HEC_URL` and `PARITY_SPLUNK_MGMT_URL`.
- The collector binary. Build it with `make otelcol` from the repo root.

Pointing at an existing Splunk is what lets the suite run where it cannot start
its own — a GitHub-hosted Windows runner runs Windows containers and cannot host
the Linux image — and it takes the pull and boot out of the local loop.

`make update-goldens` additionally downloads a UF, so it needs network access and
a platform the pinned tarball is published for (Linux amd64/arm64, macOS).

Environment variables:

| Variable                 | Purpose                                                                     |
| ------------------------ | ---------------------------------------------------------------------------- |
| `PARITY_OTELCOL_BIN`     | collector binary. Defaults to `../../bin/otelcol`.                          |
| `PARITY_UF_DIR`          | UF install root. Set by `make update-goldens`.                             |
| `PARITY_SPLUNK_HEC_URL`  | HEC base URL of an existing Splunk, e.g. `https://splunk:8088`. Setting it skips the container. |
| `PARITY_SPLUNK_MGMT_URL` | Management/search base URL, e.g. `https://splunk:8089`. Required alongside the above. |
| `PARITY_SPLUNK_TOKEN`    | A HEC token the instance accepts. Required when external, and must be >= 36 chars. |
| `PARITY_SPLUNK_PASSWORD` | Search password. Required when external.                                   |
| `PARITY_SPLUNK_USER`     | Search user. Defaults to `admin`.                                           |
| `PARITY_SPLUNK_IMAGE`    | override the Splunk image (e.g. a locally cached tag). Container path only. |

Setting `PARITY_SPLUNK_HEC_URL` or `PARITY_SPLUNK_MGMT_URL` without the other is
an error, not a fallback to starting a container: both are required together,
both need an explicit `http://` or `https://` scheme, and the credentials must be
able to create indexes.

Run:

```sh
cd tests/parity
go test -v -timeout 30m .
```

On the container path, the first run pulls the Splunk image (~2.5GB) and boots
it, so allow a few minutes. CI replays on every PR via
`.github/workflows/parity-test.yml`.

Against an existing Splunk:

```sh
PARITY_SPLUNK_HEC_URL=https://localhost:8088 \
PARITY_SPLUNK_MGMT_URL=https://localhost:8089 \
PARITY_SPLUNK_TOKEN=11111111-1111-1111-1111-111111111111 \
PARITY_SPLUNK_PASSWORD=<admin password> \
go test -v -timeout 30m .
```

## Status

The suite replays the collector against UF-derived goldens on the fields each
case selects. The next milestone is adding cases and covering typed values, since
`Fields` values are strings today.
