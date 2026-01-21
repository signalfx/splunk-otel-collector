# TA v2 Package Tests

This directory contains Go tests that validate the structure and contents of the TA v2 packages.

## Running the Tests

**Important**: You must build the packages before running the tests.

From the `packaging/ta-v2` directory:

```bash
# Build the collector binaries first (from repo root)
cd ../../
make GOOS=linux GOARCH=amd64 otelcol
make GOOS=windows GOARCH=amd64 otelcol

# Copy binaries to ta-v2 assets (from repo root)
install -D bin/otelcol_linux_amd64 packaging/ta-v2/assets/linux_x86_64/bin/Splunk_TA_OTel_Collector
install -D bin/otelcol_windows_amd64 packaging/ta-v2/assets/windows_x86_64/bin/Splunk_TA_OTel_Collector.exe

# Build all packages
cd packaging/ta-v2
make all-packages

# Run the tests
make test
```

Or run directly from the tests directory:

```bash
cd tests
go test -v ./...
```

## Test Coverage

The tests validate:

1. **Package Sizes**: OS-specific packages (linux/windows) are smaller than the multi-OS package
2. **Linux Package**: Contains only Linux binaries, excludes Windows binaries
3. **Windows Package**: Contains only Windows binaries, excludes Linux binaries
4. **Multi-OS Package**: Contains both Linux and Windows binaries
5. **Mandatory Files**: All packages contain required configuration files:
   - `default/app.conf`
   - `default/inputs.conf`
   - `README/inputs.conf.spec`
   - `configs/agent_config.yaml`

## Test Output

The tests provide detailed logging including:
- Package sizes in bytes
- Number of entries in each archive
- First 10 entries of each package (for debugging)
- Missing or unexpected files

## Prerequisites

The tests require:
- Go 1.23 or later
- Built packages in `../out/distribution/`
