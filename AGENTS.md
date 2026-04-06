# Repository Guidelines

## Project Structure & Module Organization
- `cmd/otelcol` builds the collector binary. `cmd/discoverybundler` and `cmd/ta-inputs-from-schema` provide supporting CLIs.
- `internal/` holds private implementation details, config sources, and agent utilities. Keep non-public code here.
- `pkg/` contains reusable receivers, processors, extensions, and modular input support. Some packages are nested Go modules.
- `tests/` contains integration suites, test utilities in `tests/testutils`, and scenario data under `testdata/`.
- `docs/`, `examples/`, and `deployments/` contain docs, runnable configs, and packaging assets. Local build output goes to `bin/`.

## Build, Test, and Development Commands
- `make install-tools` installs contributor tooling such as `golangci-lint`, `gofumpt`, `goimports`, and `chloggen`.
- `make otelcol` runs generation and builds the collector into `bin/otelcol`.
- `make test` runs unit tests across Go packages with race detection enabled by default.
- `make integration-test` or `make integration-test-<name>` runs Docker-backed suites from `tests/`.
- `make lint` checks licenses, spelling, and `golangci-lint`; `make fmt` applies the expected formatters and import ordering.
- `make chlog-new` and `make chlog-validate` create and validate user-facing changelog entries.

## Coding Style & Naming Conventions
- Follow standard Go style; let `make fmt` handle `gofumpt`, `gofmt -s`, and `goimports`.
- Use lower-case package names, `CamelCase` for exported identifiers, and descriptive names such as `warning_provider_test.go`.
- Keep fixtures in `testdata/`, and avoid committing generated binaries, coverage output, or local environment artifacts.

## Testing Guidelines
- Keep unit tests next to the code in `*_test.go`; place environment-dependent tests under `tests/`.
- Prefer targeted test runs while iterating, then finish with `make test` and any relevant integration target.
- No repository-wide coverage gate is documented, but new behavior should include automated tests and a validation note in the PR.

## Commit & Pull Request Guidelines
- Recent commits use short, imperative subjects, often with an issue reference, for example `Replace usage of issue creation action with gh command (#7391)`. Prefixes like `[chore]` are common.
- Keep each PR focused on one change and, when possible, under roughly 500 lines.
- Complete `.github/pull_request_template.md`: include a description, Splunk idea link, testing summary, and documentation updates.
- Add a `.chloggen/*.yaml` entry for user-visible behavior, config, or default-setting changes.

## Security & Configuration Tips
- Do not commit access tokens, realms, or machine-specific credentials. Start from `cmd/otelcol/config` or `examples/` and inject secrets through environment variables or supported config providers.
- Review docs and deployment assets before changing default ports, packaging flows, or platform-specific installers.
