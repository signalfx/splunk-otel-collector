# AGENTS instructions

## AI-assisted contributions

Follow the AI-assisted contribution policy in
[CONTRIBUTING.md](CONTRIBUTING.md#ai-assisted-contributions). In particular: never post boilerplate or auto-generated replies to review comments
(including to automated reviewers like Copilot) — address feedback in code or respond in
your own words. Disclose significant AI assistance with an `Assisted-by:` commit message
trailer, never `Co-authored-by:` (it breaks EasyCLA).

## Validation

Before considering any Go change complete, run `go build ./...` from the repo root and confirm
it exits clean. Do this for every change, not just ones that touch `cmd/otelcol` or `pkg/`.

- Use `go build ./...` or `go run ./cmd/otelcol -- <args>` (package path). Do not run
  `go run cmd/otelcol/main.go`: `cmd/otelcol` splits platform-specific code (e.g. `run`) across
  `main_others.go`/`main_windows.go`/`main_fipsonly.go`, and `go run` on a single file only
  compiles that file, so it fails with `undefined: run` even when the package builds fine.
- After `go build ./...` passes, run the relevant `go test ./...` packages for the area you
  changed.
- `go build ./...` only proves this module builds; it does not prove a package can be
  imported from *outside* this module. This module's `go.mod` has local `replace` directives
  (e.g. `github.com/signalfx/signalfx-agent => ./internal/signalfx-agent`) that only apply
  when this module is the main module being built — an external importer's `go.mod` doesn't
  get them. If you add a public package meant for external import (like `pkg/components`),
  keep it free of any import chain that reaches a locally-replaced module, and if unsure,
  validate by building it from a throwaway module elsewhere that only has a
  `replace github.com/signalfx/splunk-otel-collector => <local path>` pointing at this repo.

## PR descriptions

Keep PR descriptions concise. Do not repeat yourself. Do not add unimportant or unrelated text.

- Skip section titles (Summary, Test plan, etc.) when the description is short — just write the prose.
- Skip the Test plan for simple changes (typos, copy edits, single-line fixes, doc tweaks).
- Do not list detailed changes (file-by-file breakdowns, bullet lists of every edit) when the main description already conveys what changed and why. Only enumerate changes when the diff is large or non-obvious enough that a reviewer needs the map.
