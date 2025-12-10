# Running Puppet Tests Locally

This guide explains how to run the Puppet deployment tests locally before pushing to CI.

## Prerequisites

1. **Docker** - Must be installed and running
   - **For ARM64 Mac users**: Ensure Docker Desktop has "Use Rosetta for x86/amd64 emulation on Apple Silicon" enabled in Settings > General
   - Alternatively, Docker Desktop should automatically handle multi-platform builds
2. **Python 3** (3.13 recommended, as used in CI)
3. **pip** and **virtualenv**

## Setup Steps

### 1. Create and activate a virtual environment

```bash
cd /path/to/splunk-otel-collector
virtualenv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

### 2. Install Python dependencies

```bash
pip install -r packaging/tests/requirements.txt
```

### 3. Build the required packages

The tests require DEB and RPM packages to be present in the `dist/` directory. Build them first:

```bash
# Build DEB package for amd64
make deb-package ARCH=amd64 VERSION=0.0.1-local

# Build RPM package for amd64
make rpm-package ARCH=amd64 VERSION=0.0.1-local
```

**Note**: If you already have binaries and bundles built, you can skip compilation:

```bash
make deb-package ARCH=amd64 VERSION=0.0.1-local SKIP_COMPILE=true SKIP_BUNDLE=true
make rpm-package ARCH=amd64 VERSION=0.0.1-local SKIP_COMPILE=true SKIP_BUNDLE=true
```

The packages will be created in `dist/`:
- `dist/splunk-otel-collector_0.0.1-local_amd64.deb`
- `dist/splunk-otel-collector-0.0.1-local.x86_64.rpm`

## Running Tests

### Run all Puppet tests

```bash
# From the repo root, with venv activated
pytest -s --verbose packaging/tests/deployments/puppet/puppet_test.py
```

### Run tests for a specific distro

```bash
# For Ubuntu Jammy (DEB)
pytest -s --verbose -k "ubuntu-jammy" packaging/tests/deployments/puppet/puppet_test.py

# For CentOS 8 (RPM)
pytest -s --verbose -k "centos-8" packaging/tests/deployments/puppet/puppet_test.py
```

### Run a specific test function

```bash
# Run only test_puppet_default
pytest -s --verbose -k "test_puppet_default" packaging/tests/deployments/puppet/puppet_test.py

# Run only test_puppet_with_custom_vars
pytest -s --verbose -k "test_puppet_with_custom_vars" packaging/tests/deployments/puppet/puppet_test.py
```

### Run tests for a specific Puppet release

```bash
# Set PUPPET_RELEASE environment variable
export PUPPET_RELEASE=7
pytest -s --verbose packaging/tests/deployments/puppet/puppet_test.py

# Or test multiple releases
export PUPPET_RELEASE=7,8
pytest -s --verbose packaging/tests/deployments/puppet/puppet_test.py
```

### Run tests with specific markers

```bash
# Run only DEB distro tests
pytest -s --verbose -m deb packaging/tests/deployments/puppet/puppet_test.py

# Run only RPM distro tests
pytest -s --verbose -m rpm packaging/tests/deployments/puppet/puppet_test.py

# Run instrumentation tests
pytest -s --verbose -m instrumentation packaging/tests/deployments/puppet/puppet_test.py

# Exclude instrumentation tests
pytest -s --verbose -m "not instrumentation" packaging/tests/deployments/puppet/puppet_test.py
```

### Combine filters

```bash
# Run test_puppet_default for Ubuntu Jammy with Puppet 7, excluding instrumentation
pytest -s --verbose -k "test_puppet_default and ubuntu-jammy and not instrumentation" \
  -m deb \
  packaging/tests/deployments/puppet/puppet_test.py
```

## Available Test Distros

### DEB Distros
- `debian-bullseye`
- `ubuntu-focal`
- `ubuntu-jammy`

### RPM Distros
- `amazonlinux-2023`
- `centos-8`
- `centos-9`
- `opensuse-15`
- `oraclelinux-8`

## Common pytest Options

- `-s` - Don't capture output (see print statements)
- `-v` or `--verbose` - Verbose output
- `-k EXPRESSION` - Run tests matching the expression
- `-m MARKER` - Run tests with specific markers
- `--last-failed` - Re-run only the tests that failed last time
- `-x` - Stop on first failure
- `--pdb` - Drop into debugger on failure

## Example: Quick Test Run

```bash
# 1. Setup (one time)
cd /path/to/splunk-otel-collector
virtualenv venv
source venv/bin/activate
pip install -r packaging/tests/requirements.txt

# 2. Build packages (whenever code changes)
make deb-package ARCH=amd64 VERSION=0.0.1-local SKIP_COMPILE=true SKIP_BUNDLE=true
make rpm-package ARCH=amd64 VERSION=0.0.1-local SKIP_COMPILE=true SKIP_BUNDLE=true

# 3. Run a quick test
pytest -s --verbose -k "test_puppet_default and ubuntu-jammy" \
  packaging/tests/deployments/puppet/puppet_test.py
```

## Troubleshooting

### Packages not found

Ensure packages exist in `dist/`:
```bash
ls -la dist/splunk-otel-collector*
```

### Docker issues

Ensure Docker is running:
```bash
docker ps
```

### Architecture mismatch (ARM64 Mac)

If you see "exec format error" when running tests on an ARM64 Mac:

1. **Enable Docker Desktop emulation**:
   - Open Docker Desktop
   - Go to Settings > General
   - Enable "Use Rosetta for x86/amd64 emulation on Apple Silicon"
   - Restart Docker Desktop

2. **Verify platform support**:
   ```bash
   docker buildx ls
   ```

3. **Test with explicit platform** (if needed):
   ```bash
   docker run --platform linux/amd64 ubuntu:22.04 echo "test"
   ```

### Puppet version compatibility

- **Puppet 6**: Not supported on Ubuntu Jammy (22.04). Tests will automatically skip.
- **Puppet 7+**: Supported on all test distros.

### Permission issues

On Linux, you may need to run Docker commands with sudo or add your user to the docker group.

### Test timeouts

Tests can take several minutes. Increase timeout if needed:
```bash
pytest --timeout=600 packaging/tests/deployments/puppet/puppet_test.py
```

## CI vs Local Differences

- **CI**: Uses matrix strategy to test all distros and Puppet releases
- **Local**: You can selectively test specific combinations
- **CI**: Downloads artifacts from previous jobs
- **Local**: You must build packages manually before testing

## Next Steps

After validating locally, you can push to CI which will run the full test matrix automatically.

