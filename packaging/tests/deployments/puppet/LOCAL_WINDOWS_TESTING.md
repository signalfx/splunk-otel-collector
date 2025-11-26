# Local Windows Testing Guide for Puppet Deployment

This guide explains how to build Windows MSI artifacts locally and test the Puppet deployment with them.

## Prerequisites

### On Windows Build Machine

1. **Git Bash** - For running build scripts
2. **Go 1.21+** - For building the collector binary
3. **WiX Toolset 3.14.0.8606** - For building MSI
   - Download from: https://github.com/wixtoolset/wix3/releases/tag/wix3141rtm
   - **Important**: Use exactly version 3.14.0.8606 (not later versions)
4. **Visual Studio** or **Build Tools for Visual Studio** - For building custom actions
5. **PowerShell** - For building agent bundle

### On Test Machine (Windows)

1. **Python 3.8+** with pytest
2. **Chocolatey** - Package manager for Windows
3. **Puppet Agent** - Will be installed automatically by test script

## Step-by-Step Build Process

### 1. Build Windows Binary

From the repository root on Windows (or cross-compile from Linux/Mac):

```bash
# On Windows with Git Bash or WSL
make binaries-windows_amd64

# Or cross-compile from Linux/Mac
GOOS=windows GOARCH=amd64 make binaries-windows_amd64
```

This creates: `bin/otelcol_windows_amd64.exe`

### 2. Build Agent Bundle (Windows Required)

On a Windows machine, run:

```powershell
# From repository root
cd packaging/bundle/scripts/windows
.\make.ps1 bundle
```

This creates: `dist/agent-bundle_windows_amd64.zip`

**Note**: This step must be done on Windows as it involves platform-specific dependencies.

### 3. Build MSI (Windows Required)

On a Windows machine with Git Bash and WiX Toolset installed:

```bash
# From repository root in Git Bash
cd packaging/msi

# Build with auto-detected version
./build.sh

# Or specify a version (must be in format X.Y.Z or X.Y.Z.W)
./build.sh 0.0.1-local
```

This creates: `dist/splunk-otel-collector-<version>-amd64.msi`

**Important**: 
- The MSI filename format is: `splunk-otel-collector-<version>-amd64.msi`
- Version must be in Windows-compatible format (e.g., `0.0.1`, `1.2.3.4`)
- Pre-release versions are automatically converted (e.g., `v0.130.1-rc.0` → `0.130.1.0`)

### 4. Verify Artifacts

Check that all required artifacts were built:

```bash
python packaging/tests/deployments/puppet/local_windows_test.py --check
```

Expected output:
```
✓ Found: Windows Binary (bin/otelcol_windows_amd64.exe)
✓ Found: Agent Bundle (dist/agent-bundle_windows_amd64.zip)
✓ Found: MSI files (1 file(s))
    - splunk-otel-collector-0.0.1-amd64.msi (version: 0.0.1)
```

## Running Tests with Local Artifacts

### 1. Start Local MSI Server

In one terminal, start the HTTP server to serve your local MSI:

```bash
python packaging/tests/deployments/puppet/local_windows_test.py --serve
```

This starts a server on `http://localhost:8000` serving files from the `dist/` directory.

The server will display the exact command to run the tests:

```
Starting HTTP server on port 8000...
Serving files from: /path/to/splunk-otel-collector/dist

To run puppet tests, use:

  WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:8000 \
  pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default
```

### 2. Run Puppet Tests

In another terminal on the Windows test machine, run:

```bash
# Set the version to match your MSI file
WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:8000 \
pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default

# Or run the custom vars test
WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:8000 \
pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_custom_vars

# Or run both tests
WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:8000 \
pytest -v packaging/tests/deployments/puppet/puppet_test.py -k test_win_puppet
```

### 3. Using Custom Port

If port 8000 is already in use:

```bash
# Start server on different port
python packaging/tests/deployments/puppet/local_windows_test.py --serve --port 9000

# Run tests with custom port
WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:9000 \
pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default
```

## Environment Variables Reference

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `WIN_COLLECTOR_VERSION` | Version of the collector to test | `123.456.789` (will fail) | `0.0.1` |
| `LOCAL_MSI_SERVER` | URL of local MSI server | Official Splunk CDN | `http://localhost:8000` |
| `PUPPET_RELEASE` | Puppet agent version to install | `latest` | `7.30.0` |

## Troubleshooting

### MSI Build Fails

**Error**: `agent-bundle_windows_amd64.zip not found`
- **Solution**: Build the agent bundle first (step 2 above)

**Error**: `candle.exe not found`
- **Solution**: Install WiX Toolset 3.14.0.8606 and ensure it's in PATH

**Error**: `Unexpected candle.exe version`
- **Solution**: Use exactly version 3.14.0.8606 (later versions have issues)

### Test Fails to Download MSI

**Error**: Puppet can't download MSI from local server
- **Solution**: Ensure local server is running and accessible
- Check firewall settings
- Verify the URL: `http://localhost:8000/splunk-otel-collector-<version>-amd64.msi`
- Test with curl: `curl -I http://localhost:8000/splunk-otel-collector-0.0.1-amd64.msi`

### Version Mismatch

**Error**: Test looking for wrong version
- **Solution**: Ensure `WIN_COLLECTOR_VERSION` matches the MSI filename exactly
- MSI filename: `splunk-otel-collector-0.0.1-amd64.msi` → use `WIN_COLLECTOR_VERSION=0.0.1`

### Puppet Installation Fails

**Error**: `choco not installed!`
- **Solution**: Install Chocolatey: https://chocolatey.org/install

## Quick Reference

### Complete Build & Test Flow

```bash
# On Windows machine:

# 1. Build all artifacts
make binaries-windows_amd64
cd packaging/bundle/scripts/windows && ./make.ps1 bundle && cd ../../../../
./packaging/msi/build.sh 0.0.1

# 2. Verify artifacts
python packaging/tests/deployments/puppet/local_windows_test.py --check

# 3. In Terminal 1: Start MSI server
python packaging/tests/deployments/puppet/local_windows_test.py --serve

# 4. In Terminal 2: Run tests
WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:8000 \
pytest -v packaging/tests/deployments/puppet/puppet_test.py -k test_win_puppet
```

## Additional Notes

- The local MSI server is a simple HTTP server and doesn't support HTTPS
- Puppet module will download and cache the MSI in `C:\Windows\Temp\`
- After successful test, the collector will be installed in `C:\Program Files\Splunk\OpenTelemetry Collector\`
- Test cleanup may require manual uninstallation of the collector

## Getting Help

For issues specific to:
- **Building**: Check the build logs and verify prerequisites
- **Testing**: Run with `-v -s` for verbose pytest output
- **Puppet**: Check `C:\ProgramData\PuppetLabs\puppet\var\log\` for Puppet logs

## Related Documentation

- [Main README](../../../../README.md)
- [Puppet Deployment Guide](../../../../deployments/puppet/README.md)
- [MSI Build Documentation](../../../msi/README.md)

