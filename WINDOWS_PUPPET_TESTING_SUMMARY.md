# Summary: Windows Puppet Testing with Local Build

This document summarizes the changes made to enable testing Puppet deployment with locally built Windows artifacts.

## Changes Made

### 1. Test Infrastructure (`packaging/tests/deployments/puppet/puppet_test.py`)

**Added support for local MSI server:**
- New environment variable `LOCAL_MSI_SERVER` to override default MSI repository URL
- Defaults to official CDN: `https://dl.signalfx.com/splunk-otel-collector/msi/release`
- For local testing: `LOCAL_MSI_SERVER=http://localhost:8000`

**Updated test functions:**
- `test_win_puppet_default()`: Now passes `win_repo_url` parameter to Puppet
- `test_win_puppet_custom_vars()`: Uses new `WIN_CUSTOM_VARS_CONFIG` template with repo URL support

**New configuration template:**
```python
WIN_CUSTOM_VARS_CONFIG = string.Template(
    """
    class { splunk_otel_collector:
        ...
        win_repo_url => '$win_repo_url',
        ...
    }
    """
)
```

### 2. Local Testing Helper (`packaging/tests/deployments/puppet/local_windows_test.py`)

**New Python script** with the following features:

- **`--check`**: Verify all required artifacts exist
  ```bash
  python packaging/tests/deployments/puppet/local_windows_test.py --check
  ```

- **`--serve`**: Start HTTP server to serve MSI files
  ```bash
  python packaging/tests/deployments/puppet/local_windows_test.py --serve [--port 8000]
  ```

- **Auto-detection**: Finds MSI files and extracts version information
- **Instructions**: Displays exact commands to run tests

### 3. Build Helper Script (`packaging/msi/build-for-testing.sh`)

**New bash script** that:
- Validates prerequisites (Windows binary, agent bundle)
- Runs the MSI build process
- Provides clear instructions for testing
- Shows the exact test command with detected version

Usage:
```bash
./packaging/msi/build-for-testing.sh [version]
# Default version: 0.0.1-local
```

### 4. Comprehensive Documentation (`packaging/tests/deployments/puppet/LOCAL_WINDOWS_TESTING.md`)

**Complete guide** covering:
- Prerequisites for Windows build machine
- Step-by-step build instructions
- Testing workflow
- Environment variables reference
- Troubleshooting common issues
- Quick reference commands

## Usage Workflow

### Quick Start (Windows Machine)

```bash
# 1. Build Windows binary
make binaries-windows_amd64

# 2. Build agent bundle
cd packaging/bundle/scripts/windows
./make.ps1 bundle
cd ../../../../

# 3. Build MSI with helper script
./packaging/msi/build-for-testing.sh 0.0.1

# 4. Check artifacts
python packaging/tests/deployments/puppet/local_windows_test.py --check

# 5. Terminal 1: Start MSI server
python packaging/tests/deployments/puppet/local_windows_test.py --serve

# 6. Terminal 2: Run tests
WIN_COLLECTOR_VERSION=0.0.1 LOCAL_MSI_SERVER=http://localhost:8000 \
pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default
```

## Key Features

### 1. **No External Dependencies**
- Tests can run completely offline with locally built artifacts
- No need to upload to CDN for testing

### 2. **Version Flexibility**
- Use any version string for testing (e.g., `0.0.1-local`, `dev-build`, etc.)
- Automatic version detection from MSI filename

### 3. **Simple HTTP Server**
- Pure Python, no additional dependencies
- Serves files from `dist/` directory
- Easy to debug (see all download requests)

### 4. **Backward Compatible**
- Without `LOCAL_MSI_SERVER`, tests use official CDN (existing behavior)
- No changes required for CI/CD pipelines

## Environment Variables

| Variable | Purpose | Default | Example |
|----------|---------|---------|---------|
| `WIN_COLLECTOR_VERSION` | MSI version to install | `123.456.789` | `0.0.1` |
| `LOCAL_MSI_SERVER` | MSI repository URL | Official CDN | `http://localhost:8000` |
| `PUPPET_RELEASE` | Puppet agent version | `latest` | `7.30.0` |

## File Structure

```
splunk-otel-collector/
├── packaging/
│   ├── msi/
│   │   ├── build.sh                      # Original MSI build script
│   │   └── build-for-testing.sh          # New: Helper for testing builds
│   └── tests/
│       └── deployments/
│           └── puppet/
│               ├── puppet_test.py        # Modified: Support LOCAL_MSI_SERVER
│               ├── local_windows_test.py # New: Test helper script
│               └── LOCAL_WINDOWS_TESTING.md  # New: Documentation
└── WINDOWS_PUPPET_TESTING_SUMMARY.md     # This file
```

## Benefits

1. **Faster Development Cycle**
   - No need to upload artifacts to CDN
   - Test changes immediately
   - Iterate quickly on MSI packaging

2. **Isolated Testing**
   - Test specific versions
   - No interference with production artifacts
   - Reproducible builds

3. **Better Debugging**
   - See exactly which files are being downloaded
   - Easily swap different MSI versions
   - Test error conditions

4. **CI/CD Ready**
   - Can be integrated into CI pipelines
   - Automated testing with custom builds
   - No special infrastructure needed

## Testing Checklist

Before running tests, ensure:

- [ ] Windows binary built (`bin/otelcol_windows_amd64.exe`)
- [ ] Agent bundle built (`dist/agent-bundle_windows_amd64.zip`)
- [ ] MSI built (`dist/splunk-otel-collector-*-amd64.msi`)
- [ ] Local MSI server running (Terminal 1)
- [ ] `WIN_COLLECTOR_VERSION` matches MSI filename
- [ ] `LOCAL_MSI_SERVER` points to server (e.g., `http://localhost:8000`)

## Troubleshooting

### Common Issues

1. **"No MSI files found"**
   - Run: `./packaging/msi/build-for-testing.sh`

2. **"agent-bundle_windows_amd64.zip not found"**
   - Run: `packaging/bundle/scripts/windows/make.ps1 bundle`

3. **"candle.exe not found"**
   - Install WiX Toolset 3.14.0.8606

4. **Puppet can't download MSI**
   - Check server is running
   - Verify URL: `curl -I http://localhost:8000/splunk-otel-collector-0.0.1-amd64.msi`
   - Check firewall settings

## Next Steps

For production testing:
1. Build artifacts on Windows machine
2. Use helper scripts for validation
3. Run local tests
4. Document any issues found
5. Iterate until tests pass

## Additional Resources

- Full documentation: `packaging/tests/deployments/puppet/LOCAL_WINDOWS_TESTING.md`
- Puppet deployment: `deployments/puppet/README.md`
- Main README: `README.md`

---

**Questions or issues?** Check the troubleshooting section in LOCAL_WINDOWS_TESTING.md or review the build logs for specific errors.

