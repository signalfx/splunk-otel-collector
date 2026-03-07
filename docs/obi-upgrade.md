# Upgrading the OBI receiver to a new version

This guide covers how to upgrade the OBI (OpenTelemetry eBPF Instrumentation)
receiver to a new upstream release.

## Background

OBI source is not committed to this repo. Instead, it is fetched at build time
from an official GitHub release tarball that includes all pre-generated BPF
files. The tarball URL follows this pattern:

```
https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/download/vX.Y.Z/obi-vX.Y.Z-source-generated.tar.gz
```

The `OBI_VERSION` variable in the `Makefile` (currently `v0.6.0`) controls which
version is used everywhere: `make fetch-obi`, the CI action, and the `go.mod`
`require` directive.

## Step-by-step upgrade procedure

### 1. Confirm the new release tarball exists

Visit the OBI releases page and confirm the `obi-vX.Y.Z-source-generated.tar.gz`
asset is present:

```
https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/tag/vX.Y.Z
```

> **Note:** Only tarballs named `*-source-generated.tar.gz` include the
> pre-generated BPF files required to build without a BPF toolchain. Plain
> source archives from GitHub's automatic "Source code" assets do NOT include
> those files and will not work.

### 2. Update `OBI_VERSION` in `Makefile`

```makefile
# Before:
OBI_VERSION?=v0.6.0

# After:
OBI_VERSION?=vX.Y.Z
```

### 3. Update `go.mod`

```
# Before:
go.opentelemetry.io/obi v0.6.0

# After:
go.opentelemetry.io/obi vX.Y.Z
```

The `replace` directive (`go.opentelemetry.io/obi => ./third_party/opentelemetry-ebpf-instrumentation`)
does not need to change.

### 4. Fetch the new OBI source

Remove any previously extracted source and fetch the new version:

```bash
rm -rf third_party/opentelemetry-ebpf-instrumentation
make fetch-obi OBI_VERSION=vX.Y.Z
```

Or simply run `make fetch-obi` after updating `OBI_VERSION` in the Makefile.

### 5. Update Go dependencies

```bash
go mod tidy
```

This updates `go.sum` and resolves any transitive dependency changes introduced
by the new OBI version.

### 6. Check for API changes

If OBI changed its collector API (e.g., `go.opentelemetry.io/obi/collector`),
update `internal/components/components.go` accordingly.

Review the OBI changelog for breaking changes:

```
https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/main/CHANGELOG.md
```

### 7. Build and test

```bash
make otelcol
go test ./internal/components/...
```

### 8. Update the CI action default version

The `fetch-obi` action at `.github/actions/fetch-obi/action.yml` has a default
input version. Update it to match:

```yaml
inputs:
  obi-version:
    description: OBI version to fetch (e.g., v0.6.0)
    default: vX.Y.Z   # ← update this
```

### 9. Commit

Stage and commit all changes:

```bash
git add Makefile go.mod go.sum .github/actions/fetch-obi/action.yml
git commit -m "Bump OBI receiver to vX.Y.Z"
```

## Checklist

- [ ] `OBI_VERSION` updated in `Makefile`
- [ ] `go.opentelemetry.io/obi vX.Y.Z` in `go.mod`
- [ ] `third_party/opentelemetry-ebpf-instrumentation/` re-fetched from new tarball
- [ ] `go mod tidy` run, `go.sum` updated
- [ ] Default version in `.github/actions/fetch-obi/action.yml` updated
- [ ] `internal/components/components.go` updated if OBI API changed
- [ ] `make otelcol` builds cleanly
- [ ] `go test ./internal/components/...` passes

## Troubleshooting

### `go build` fails with "no matching files found"

The extracted tarball is missing the generated `*_bpfel.go` files. This means:
- The tarball used is a plain source archive (not `*-source-generated.tar.gz`), **OR**
- The extraction failed partially.

Solution: remove `third_party/opentelemetry-ebpf-instrumentation/` and re-run
`make fetch-obi`.

### `make fetch-obi` fails with a 404

The `obi-vX.Y.Z-source-generated.tar.gz` asset does not exist for that release.
Check the OBI releases page. If the tarball is genuinely missing, open an issue
upstream at <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation>.

### Build cache issues after upgrade

Run `go clean -cache` to clear the Go build cache if you see stale compilation
errors after an OBI upgrade.
