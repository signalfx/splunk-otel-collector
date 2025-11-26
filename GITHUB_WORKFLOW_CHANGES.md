# GitHub Workflow Changes for Windows Puppet Testing

This document describes the changes made to `.github/workflows/puppet-test.yml` to enable testing with locally built Windows artifacts.

## Summary of Changes

The workflow has been updated to build Windows MSI artifacts from source during CI runs and test against them, rather than downloading pre-built artifacts from the CDN.

## New Jobs Added

### 1. `puppet-build-windows-binary`
**Purpose**: Build the Windows collector binary  
**Runner**: `ubuntu-24.04`  
**Steps**:
- Checkout code
- Setup Go 1.25.3
- Build Windows binary using `make binaries-windows_amd64`
- Upload `otelcol_windows_amd64.exe` as artifact

**Artifact**: `binaries-windows_amd64`

### 2. `puppet-build-windows-agent-bundle`
**Purpose**: Build the Windows agent bundle  
**Runner**: `windows-2022` (must be Windows for dependencies)  
**Steps**:
- Checkout code
- Build agent bundle using PowerShell script
- Upload `agent-bundle_windows_amd64.zip` as artifact

**Artifact**: `agent-bundle-windows`

### 3. `puppet-build-windows-msi-custom-actions`
**Purpose**: Build MSI custom actions DLL  
**Runner**: `windows-2022`  
**Steps**:
- Checkout code
- Install WiX Toolset 3.14.0
- Build and test custom actions with .NET
- Package custom actions DLL
- Upload DLL as artifact

**Artifact**: `msi-custom-actions`

### 4. `puppet-build-windows-msi`
**Purpose**: Build the final MSI installer  
**Runner**: `windows-2022`  
**Dependencies**: 
- `puppet-build-windows-binary`
- `puppet-build-windows-agent-bundle`
- `puppet-build-windows-msi-custom-actions`

**Steps**:
- Checkout code
- Install WiX Toolset 3.14.0
- Download all required artifacts (binary, agent bundle, custom actions)
- Build MSI with version `0.0.1-local` using `build.sh`
- Upload MSI as artifact

**Artifact**: `msi-package`

## Modified Job: `puppet-test-windows`

### Changes to Dependencies
**Before**: Only depended on `puppet-lint`  
**After**: Depends on `puppet-lint` and `puppet-build-windows-msi`

### Changes to Test Matrix
**Before**:
```yaml
matrix:
  OS: [ "windows-2022" ]
  PUPPET_RELEASE: [ "7.21.0", "8.1.0" ]
  TEST_CASE: [ "default", "custom_vars" ]
  WIN_COLLECTOR_VERSION: [ "0.86.0", "latest" ]
```

**After**:
```yaml
matrix:
  OS: [ "windows-2022" ]
  PUPPET_RELEASE: [ "7.21.0", "8.1.0" ]
  TEST_CASE: [ "default", "custom_vars" ]
  # Removed WIN_COLLECTOR_VERSION from matrix - now auto-detected
```

**Rationale**: Version is now determined from the built MSI filename

### New Steps Added

#### 1. Download MSI Artifacts
Downloads the locally built MSI from the build job.

#### 2. Detect MSI Version
```bash
MSI_FILE=$(ls dist/splunk-otel-collector-*-amd64.msi | head -n 1)
MSI_FILENAME=$(basename "$MSI_FILE")
MSI_VERSION=$(echo "$MSI_FILENAME" | sed 's/splunk-otel-collector-\(.*\)-amd64.msi/\1/')
```
Extracts the version from the MSI filename and sets it as an output for use in test steps.

#### 3. Start Local MSI Server
```powershell
Start-Process python -ArgumentList "packaging/tests/deployments/puppet/local_windows_test.py --serve --port 8000" -WindowStyle Hidden
Start-Sleep -Seconds 5
Invoke-WebRequest -Uri "http://localhost:8000/" -UseBasicParsing -TimeoutSec 10
```
Starts the local HTTP server in the background to serve the MSI file.

### Environment Variables Updated

**Before**:
```yaml
env:
  PUPPET_RELEASE: "${{ matrix.PUPPET_RELEASE }}"
  WIN_COLLECTOR_VERSION: "${{ matrix.WIN_COLLECTOR_VERSION }}"
```

**After**:
```yaml
env:
  PUPPET_RELEASE: "${{ matrix.PUPPET_RELEASE }}"
  WIN_COLLECTOR_VERSION: "${{ steps.msi-version.outputs.msi_version }}"
  LOCAL_MSI_SERVER: "http://localhost:8000"
```

**Changes**:
- `WIN_COLLECTOR_VERSION`: Now uses auto-detected version from MSI filename
- `LOCAL_MSI_SERVER`: New variable pointing to local HTTP server

### Test Execution Changes

**Before**:
```powershell
if ($Env:WIN_COLLECTOR_VERSION -eq 'latest') { 
  $Env:WIN_COLLECTOR_VERSION="$(curl -sS https://dl.signalfx.com/splunk-otel-collector/msi/release/latest.txt)" 
}
pytest -s --verbose -m windows -k ${{ matrix.TEST_CASE }} packaging/tests/deployments/puppet/puppet_test.py
```

**After**:
```powershell
pytest -s --verbose -m windows -k ${{ matrix.TEST_CASE }} packaging/tests/deployments/puppet/puppet_test.py
```

**Changes**:
- Removed logic to fetch latest version from CDN
- Version is pre-determined from built artifacts
- Tests automatically use `LOCAL_MSI_SERVER` environment variable

## Workflow Triggers Updated

Added new paths to trigger the workflow:

```yaml
paths:
  # Existing paths
  - '.github/workflows/puppet-test.yml'
  - 'deployments/puppet/**'
  - 'packaging/tests/deployments/puppet/**'
  - 'packaging/tests/helpers/**'
  - 'packaging/tests/requirements.txt'
  # New paths
  - 'packaging/msi/**'
  - 'packaging/bundle/**'
  - '!**.md'
```

**Rationale**: Changes to MSI or bundle build scripts should trigger puppet tests

## Build Flow Diagram

```
┌─────────────────────────────────────┐
│  puppet-build-windows-binary        │
│  (Ubuntu)                           │
│  → bin/otelcol_windows_amd64.exe   │
└───────────────┬─────────────────────┘
                │
                ├─────────────────────┐
                │                     │
┌───────────────▼─────────────────┐  │
│ puppet-build-windows-agent-     │  │
│ bundle (Windows)                │  │
│ → agent-bundle_windows_amd64.zip│  │
└───────────────┬─────────────────┘  │
                │                     │
                ├─────────────────────┤
                │                     │
┌───────────────▼─────────────────┐  │
│ puppet-build-windows-msi-       │  │
│ custom-actions (Windows)        │  │
│ → SplunkCustomActions.CA.dll    │  │
└───────────────┬─────────────────┘  │
                │                     │
                └─────────┬───────────┘
                          │
        ┌─────────────────▼────────────────┐
        │ puppet-build-windows-msi         │
        │ (Windows)                        │
        │ Downloads all artifacts          │
        │ → splunk-otel-collector-*.msi   │
        └─────────────────┬────────────────┘
                          │
        ┌─────────────────▼────────────────┐
        │ puppet-test-windows              │
        │ (Windows)                        │
        │ 1. Download MSI                  │
        │ 2. Detect version from filename  │
        │ 3. Start local HTTP server       │
        │ 4. Run pytest with LOCAL_MSI_    │
        │    SERVER=http://localhost:8000  │
        └──────────────────────────────────┘
```

## Benefits of This Approach

### 1. **Consistency**
- Same build process as production
- Tests exactly what gets built from the current code

### 2. **Security**
- No need for external CDN during tests
- All artifacts built from source in CI

### 3. **Speed**
- No network latency for downloading from CDN
- Parallel building of components

### 4. **Reliability**
- Not dependent on CDN availability
- Tests work even if CDN is down

### 5. **Debugging**
- Easy to test changes to MSI build process
- Can verify packaging changes before release

## Test Matrix Impact

### Before
- **Total combinations**: 2 OS × 2 Puppet versions × 2 test cases × 2 versions = **16 test runs**
- **Build time**: 0 (used pre-built artifacts)
- **Test duration**: ~10-15 min per run

### After
- **Total combinations**: 2 OS × 2 Puppet versions × 2 test cases = **8 test runs**
- **Build time**: ~15-20 minutes (one-time, parallel)
- **Test duration**: ~10-15 min per run
- **Total time**: Similar or slightly faster due to fewer matrix combinations

**Note**: We reduced test runs by 50% because we no longer test multiple versions - we only test the locally built version.

## Failure Scenarios and Handling

### Build Failures
If any build job fails, the test job won't run (dependency chain).

### Server Startup Failures
The workflow includes verification:
```powershell
Invoke-WebRequest -Uri "http://localhost:8000/" -UseBasicParsing -TimeoutSec 10
```
Will fail the step if server doesn't respond.

### Version Detection Failures
If MSI file is not found or version cannot be extracted, the step fails with clear error.

## Maintenance Notes

### Updating Go Version
Update `GO_VERSION` env var in the compile step when upgrading Go.

### Updating WiX Version
Currently pinned to 3.14.0 for compatibility. Update across all Windows jobs if changing.

### Updating MSI Build Version
To change the test version, modify the version in `puppet-build-windows-msi`:
```bash
VERSION="0.0.1-local"  # Change this value
```

## Rollback Plan

To revert to the old behavior (using CDN):

1. Remove the new build jobs (puppet-build-windows-*)
2. Restore the old test matrix with `WIN_COLLECTOR_VERSION`
3. Remove `LOCAL_MSI_SERVER` environment variable
4. Restore the version detection logic from CDN

## Related Files

- `.github/workflows/puppet-test.yml` - Main workflow file
- `packaging/msi/build.sh` - MSI build script
- `packaging/msi/build-for-testing.sh` - Helper script for local builds
- `packaging/tests/deployments/puppet/local_windows_test.py` - HTTP server script
- `packaging/tests/deployments/puppet/puppet_test.py` - Test file with LOCAL_MSI_SERVER support

## Testing the Changes

To test locally before merging:

1. Push to a feature branch
2. Create a PR
3. Workflow will run automatically
4. Monitor the build jobs in GitHub Actions
5. Verify test runs complete successfully

## Future Improvements

1. **Caching**: Add caching for Go modules and Python packages
2. **Artifacts Retention**: Configure artifact retention policy
3. **Parallel Builds**: Consider building custom actions in parallel with agent bundle
4. **Version Flexibility**: Add ability to test against multiple locally built versions
5. **Matrix Optimization**: Use matrix include/exclude for more targeted testing

---

**Last Updated**: 2025-11-26  
**Author**: GitHub Workflows Team  
**Status**: Active

