# Package and Installer Tests

## Setup

Install Docker and Go on your workstation. Package tests are Go tests and do not require
the Python virtualenv used by the installer tests.

## Running tests

1. To run the package tests, execute the following commands:
   ```
   cd tests
   PACKAGE_TEST_TYPE=deb go test -tags package_integration -v ./package
   ```
   Set `PACKAGE_TEST_TYPE` to `deb`, `rpm`, or `tar`. To run a narrower local
   subset, set `PACKAGE_TEST_DISTRO` and `PACKAGE_TEST_ARCH`. The package tests
   require that the packages to be tested exist locally in `<repo_base_dir>/dist`.
   See [here](../fpm/deb/README.md), [here](../fpm/rpm/README.md), and
   [here](../fpm/tar/README.md) for how to build the packages.
1. To run the installer tests, execute the following commands:
   ```
   virtualenv venv
   source venv/bin/activate  # if not already in virtualenv
   pip install -r packaging/tests/requirements.txt
   pytest [PYTEST_OPTIONS] packaging/tests/installer_test.py
   ```
   Installer tests still use pytest. Check [pytest.org](https://pytest.org) or
   run `pytest --help` to see the available pytest options.

## Running the `linux-installer-script-test` CI Workflow Locally

The [`installer-script-test.yml`](../../.github/workflows/installer-script-test.yml) workflow builds the
`splunk-otel-collector` deb/rpm package, then runs [`installer_test.py`](installer_test.py) against it in distro
containers using the [Linux Installer Script](../installer/install.sh). To reproduce a single
`linux-installer-script-test (<distro>, <arch>, <instrumentation>)` job locally (e.g.
`linux-installer-script-test (debian-bullseye, arm64, none)`):

1. Build the collector binary for the target arch (from the repo root):

   ```bash
   make binaries-linux_arm64   # or binaries-linux_amd64
   ```

   Produces `bin/otelcol_linux_<arch>`.

2. Build the deb/rpm package for the target distro's package type (`debian-bullseye` uses `deb`, see
   `packaging/tests/images/{deb,rpm}/Dockerfile.<distro>` to determine the type for other distros):

   ```bash
   make deb-package SKIP_COMPILE=true ARCH=arm64   # or rpm-package
   ```

   Produces `dist/splunk-otel-collector*.deb` (or `.rpm`).

3. Install the test dependencies:

   ```bash
   python3 -m venv .venv && source .venv/bin/activate
   pip install -r packaging/tests/requirements.txt
   ```

4. Point the tests at the locally built package and run pytest with the same `-k` filter CI uses. The
   `INSTRUMENTATION=none` matrix leg maps to `not instrumentation` (see
   [installer-script-test.yml:138-142](../../.github/workflows/installer-script-test.yml#L138-L142)), which
   selects `test_installer_default` and `test_installer_custom` (the only tests without the `instrumentation`
   marker):

   ```bash
   package_path=$(find ./dist -maxdepth 1 -name "splunk-otel-collector*.deb" | head -n 1)
   export LOCAL_COLLECTOR_PACKAGE=$(realpath "$package_path")

   python3 -u -m pytest -s --verbose \
     -k "debian-bullseye and arm64 and not instrumentation" \
     packaging/tests/installer_test.py
   ```

Notes:

- Docker must be running; tests spin up `--privileged` systemd containers defined under
  `packaging/tests/images/{deb,rpm}/Dockerfile.<distro>`.
- Run on a host matching the target `arch` to avoid needing QEMU emulation.
- `LOCAL_COLLECTOR_PACKAGE` tells the tests to install the package built in step 2 instead of downloading a
  released version. There's an equivalent `LOCAL_INSTRUMENTATION_PACKAGE` env var for the
  `splunk-otel-auto-instrumentation` package, only relevant to the `preload`/`systemd` instrumentation legs (not
  needed for the `none` leg). Check the [instructions on how to build the instrumentation package](
  ../../instrumentation/README.md) for details.
- Steps 1 and 2 must complete before running pytest, since the tests look up `dist/splunk-otel-collector*` for
  the given package type.
