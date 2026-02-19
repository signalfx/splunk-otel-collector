# Add OBI receiver to Splunk OTel Collector distribution

## Summary

Add OpenTelemetry eBPF Instrumentation (OBI) receiver support to `splunk-otel-collector`, following the integration direction discussed upstream in:

- <https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/46192>

This request is to include OBI as a built-in receiver in this distribution and establish a sustainable build path.

## Motivation

OBI provides zero-code instrumentation using eBPF and enables traces/metrics collection without app code changes. This is valuable for:

- legacy workloads,
- mixed-language fleets,
- rapid observability rollout,
- environments where SDK rollout is difficult.

## Proposed approach

### Option A-style integration

Integrate OBI as an external component in this distribution (do not copy code into this repo):

- Add OBI receiver factory to component wiring in `internal/components/components.go`.
- Keep OBI source of truth upstream (`open-telemetry/opentelemetry-ebpf-instrumentation`).
- Pin upstream via git submodule at `third_party/opentelemetry-ebpf-instrumentation`.

### Dependency model: submodule + `replace` is required

A release-based `go get go.opentelemetry.io/obi@v0.5.0` approach **does not work** and should not be pursued unless upstream fixes their module publishing. The reason:

- OBI's `.gitignore` at `v0.5.0` explicitly lists `*_bpfel.go` and `*_bpfel.o`.
- The Go module proxy follows `.gitignore` when creating release zips, so those files are absent from the downloaded module.
- Those files are pre-generated C-to-Go bindings (via `cilium/ebpf`'s `bpf2go`) that embed compiled eBPF bytecode. They are committed to git but stripped from the proxy zip.
- Without them, `go build` fails with: `pattern bpf_x86_bpfel.o: no matching files found`.

The working approach:

1. Add OBI upstream repo as a git submodule pinned to `v0.5.0` at `third_party/opentelemetry-ebpf-instrumentation`.
2. Add a `replace` directive in `go.mod` pointing to the submodule path.
3. Before building, run `make generate-obi` to produce the `*_bpfel.go` / `*_bpfel.o` files locally using `clang` and `bpf2go`.
4. CI uses `git clone --recurse-submodules` followed by `make generate-obi && make otelcol`.

```
# go.mod
go.opentelemetry.io/obi => ./third_party/opentelemetry-ebpf-instrumentation
```

The `make generate-obi` target handles tool availability checks and invokes OBI's own `obi_genfiles.go` driver with `OTEL_EBPF_GENFILES_RUN_LOCALLY=1`. Required host tools: `clang`, `llvm-strip`, `bpf2go` (all standard Linux packages).

Upstream could simplify this by removing `*_bpfel.go` and `*_bpfel.o` from their `.gitignore` and committing the generated files at each release tag, which would make the proxy zip complete and allow a plain `go get`-based dependency.

## Prototype status

A working prototype has been implemented and fully validated in this repo:

- OBI factory registered in `internal/components/components.go`
- Unit test receiver list updated in `internal/components/components_test.go`
- `go.mod` `replace` directive pointing to `third_party/opentelemetry-ebpf-instrumentation` (submodule at `v0.5.0`)
- `make generate-obi` target invokes OBI's local BPF generation (58s, requires `clang`/`llvm-strip`/`bpf2go`)
- `make otelcol` builds cleanly after generation and produces `./bin/otelcol_linux_amd64` (417M)
- Runtime validation: collector starts, OBI receiver initializes, logs `"Everything is ready. Begin running and processing data."`, proceeds to eBPF attachment (fails gracefully on unprivileged host as expected)

Validated build sequence:

```bash
git submodule update --init --recursive
make generate-obi   # ~60s, requires clang + llvm-strip + bpf2go
make otelcol        # ~40s with cache
```

Reference: `docs/obi-integration-prototype.md`

## Acceptance criteria

- `obi` receiver is discoverable and configurable in collector config.
- Linux builds include OBI receiver and pass component wiring tests.
- Build/documentation clearly defines when pre-generation is required.
- CI path is defined for selected dependency model (release or submodule).
- User-facing docs cover runtime privileges/capabilities for OBI.

## Risks / considerations

- OBI runtime is Linux/capability dependent.
- Source/submodule mode adds CI complexity and generation time.
- Version alignment between collector APIs and OBI module must be monitored.

## Open questions

1. Should we request upstream (OBI) to fix their `.gitignore` / module publishing so that a clean release-based dependency becomes viable long-term?
2. Do we want OBI enabled in all Linux build variants immediately, or behind a build tag/profile first?
3. Which CI lane should own initial OBI runtime validation (requires Linux with `CAP_BPF` etc.)?
4. Should `third_party/opentelemetry-ebpf-instrumentation` be a git submodule (recommended) or a committed vendor copy?
