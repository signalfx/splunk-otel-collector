# OpAMP Supervisor Packaging

The OpAMP Supervisor packaging module is used to build the upstream
`opampsupervisor` binary for Splunk OpenTelemetry Collector packages.

This module is intentionally separate from the main collector module so
packaging can pin and build the supervisor without copying or editing upstream
source.

## Development

Build the binary from the repository root with `make opampsupervisor`.

## Dependency updates

Update the supervisor by changing the pinned tool version, then running
`go mod tidy` in this directory:

```sh
go get -tool github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampsupervisor@vX.Y.Z
go mod tidy
```

Manual transitive dependency overrides will mean the packaged supervisor no
longer matches the exact upstream release dependency graph when built.
