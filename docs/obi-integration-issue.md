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
- Pin upstream via release tarball at `third_party/opentelemetry-ebpf-instrumentation` (gitignored, fetched on demand).

### Dependency model: release tarball + `replace`

Starting with OBI v0.6.0, OBI publishes a `obi-vX.Y.Z-source-generated.tar.gz`
artifact at each release that includes all pre-generated BPF files. This is the
working approach:

1. Download the OBI release tarball via `make fetch-obi`.
2. The tarball is extracted to `third_party/opentelemetry-ebpf-instrumentation/`
   (gitignored, not committed).
3. A `replace` directive in `go.mod` points to that local path.
4. No BPF toolchain (`clang`, `llvm-strip`, `bpf2go`) is required.

```
# go.mod
go.opentelemetry.io/obi => ./third_party/opentelemetry-ebpf-instrumentation
```

> **Why not `go get`?** OBI's `.gitignore` still excludes `*_bpfel.go` /
> `*_bpfel.o`. The Go module proxy follows `.gitignore` when creating release
> zips, so a plain `go get go.opentelemetry.io/obi@vX.Y.Z` produces a module
> zip that is missing the generated files. The `replace` + tarball approach
> works around this.

> **Historical note (v0.5.0):** Before v0.6.0, OBI did not publish a
> source-generated tarball. The only working approach was a git submodule +
> `make generate-obi` (requiring clang/bpf2go). That approach has been replaced
> by the tarball approach above.

## Prototype status

A working prototype has been implemented and fully validated in this repo:

- OBI factory registered in `internal/components/components.go`
- Unit test receiver list updated in `internal/components/components_test.go`
- `go.mod` `replace` directive pointing to `third_party/opentelemetry-ebpf-instrumentation` (OBI v0.6.0)
- `make fetch-obi` downloads the OBI v0.6.0 release tarball (no BPF toolchain required)
- `make otelcol` builds cleanly and produces `./bin/otelcol_linux_amd64` (421M)
- `go version -m ./bin/otelcol_linux_amd64` confirms `dep go.opentelemetry.io/obi v0.6.0`
- Runtime validation: collector starts, OBI receiver initializes, config validation runs correctly
- `TestDefaultComponents` passes, confirming OBI receiver is in the factory list

Validated build sequence:

```bash
make fetch-obi      # ~5s with cache (downloads ~46MB tarball first time)
make otelcol        # ~40s with warm Go build cache
```

Reference: `docs/obi-integration-prototype.md`
Upgrade guide: `docs/obi-upgrade.md`

## Acceptance criteria

- `obi` receiver is discoverable and configurable in collector config.
- Linux builds include OBI receiver and pass component wiring tests.
- Build/documentation clearly defines when pre-generation is required.
- CI path is defined for the tarball-based dependency model.
- User-facing docs cover runtime privileges/capabilities for OBI.

## Risks / considerations

- OBI runtime is Linux/capability dependent.
- Tarball fetch adds ~5s to CI (cached by OBI version after first run).
- Version alignment between collector APIs and OBI module must be monitored.

## Open questions

1. Do we want OBI enabled in all Linux build variants immediately, or behind a build tag/profile first?
2. Which CI lane should own initial OBI runtime validation (requires Linux with `CAP_BPF` etc.)?
3. Should CI automatically check that the OBI tarball exists for new OBI releases?
