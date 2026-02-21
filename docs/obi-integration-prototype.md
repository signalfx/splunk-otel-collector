# OBI integration prototype (Splunk OTel Collector)

## Summary

This prototype integrates OBI as a receiver in `splunk-otel-collector` by wiring `go.opentelemetry.io/obi/collector` into the collector component factory list.

It also adds a `make generate-obi` target to support an **Option A-style** workflow (submodule + pre-build generation) when building against OBI source instead of a released module.

## Feasibility of upstream "Option A"

Short answer: **yes, it is possible in this repository**.

What maps well from upstream Option A:
- Keep OBI in its own upstream repo.
- Add OBI into this distribution by importing the OBI receiver factory.
- Support a pre-build generation step via `make generate-obi` when using local/submodule source.

What is different here vs `collector-releases`:
- `splunk-otel-collector` is source-built and manually wires factories in `internal/components/components.go` (not OCB manifest-driven).
- For released OBI versions (for example `go.opentelemetry.io/obi v0.5.0`), a direct module dependency works for this integration prototype.

## Current prototype scope

Implemented:
- Added OBI receiver factory import and registration in `internal/components/components.go`.
- Added OBI receiver expectation to `internal/components/components_test.go`.
- Added module dependency: `go.opentelemetry.io/obi v0.5.0`.
- Added `make generate-obi` target for optional source/submodule generation workflows.

Not yet implemented (intentionally):
- CI wiring to run `make generate-obi` when building from a local OBI submodule/replace.
- Packaging/runtime docs for Linux capabilities and privileged execution expectations for OBI.
- End-to-end integration tests that exercise `obi` receiver at runtime.

## Suggested next implementation increments

1. Choose dependency model:
   - **Release-based (simpler):** keep module dependency on OBI release.
   - **Source/submodule-based (closer to upstream Option A):** add `third_party/opentelemetry-ebpf-instrumentation` submodule + `replace` in development/CI flows.
2. Add Linux-focused runtime documentation and config examples for `obi` receiver.
3. Add gated CI job for Linux that validates OBI build path and a minimal receiver startup test.
4. Add packaging checks to ensure capabilities/privileges expectations are documented for users.

## Notes

- OBI receiver is Linux-oriented. Non-Linux builds compile because OBI provides non-Linux factory behavior in its module.
- Runtime success for OBI depends on kernel capabilities/privileges and host setup, separate from compile-time integration.
