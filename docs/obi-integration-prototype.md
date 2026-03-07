# OBI integration prototype (Splunk OTel Collector)

## Summary

This prototype integrates OBI as a receiver in `splunk-otel-collector` by wiring
`go.opentelemetry.io/obi/collector` into the collector component factory list.

OBI is sourced from its official release tarball
(`obi-<version>-source-generated.tar.gz`) which includes all pre-generated BPF
files. This eliminates the need for a git submodule and for any BPF toolchain
(`clang`, `llvm-strip`, `bpf2go`) at build time.

## Why this approach

OBI's `.gitignore` excludes `*_bpfel.go` / `*_bpfel.o`; the Go module proxy
follows `.gitignore` when creating release zips, so a plain
`go get go.opentelemetry.io/obi@vX.Y.Z` doesn't work (missing generated files).
OBI publishes a separate tarball artifact at each release that includes all
generated files, which is what this integration uses.

## Dependency model

OBI source is downloaded from the GitHub release tarball and extracted to
`third_party/opentelemetry-ebpf-instrumentation/`. That directory is listed in
`.gitignore` (not committed). A `replace` directive in `go.mod` points the
`go.opentelemetry.io/obi` module at this local path:

```
# go.mod
go.opentelemetry.io/obi => ./third_party/opentelemetry-ebpf-instrumentation
```

## Developer workflow

```bash
# First time or after an OBI version upgrade:
make fetch-obi     # Downloads and extracts OBI vX.Y.Z tarball (~46MB, cached)

# Build the collector:
make otelcol       # ~40s with warm Go build cache
```

`make fetch-obi` is idempotent: re-running it when the source is already present
is a no-op.

## Current prototype status

Implemented:
- OBI receiver factory imported and registered in `internal/components/components.go`.
- OBI receiver listed in `internal/components/components_test.go`.
- `go.mod` `replace` directive pointing to `third_party/opentelemetry-ebpf-instrumentation`.
- `make fetch-obi` target: downloads the release tarball, extracts to `third_party/`.
- `.github/actions/fetch-obi` composite action: cross-platform (Linux + Windows),
  uses `actions/cache` keyed by OBI version.
- All CI workflows updated to use `fetch-obi` (no `generate-obi-bpf`, no `submodules: recursive`).

Not yet implemented (intentionally):
- End-to-end integration tests that exercise `obi` receiver at runtime.
- Packaging/runtime docs for Linux capabilities and privileged execution expectations for OBI.

## Upgrading to a new OBI version

See [obi-upgrade.md](obi-upgrade.md) for step-by-step instructions.

## Notes

- OBI receiver is Linux-only at runtime. Non-Linux builds compile because OBI
  provides no-op factory stubs in `factory_others.go`.
- Runtime success for OBI depends on kernel capabilities/privileges and host
  setup, separate from compile-time integration.
- The tarball is cached locally in `.local/` (gitignored). Re-running `make fetch-obi`
  after clearing `.local/` re-downloads it.
