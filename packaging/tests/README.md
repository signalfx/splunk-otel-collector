# Package and Installer Tests

## Setup

Install Docker and Go on your workstation.

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
   cd tests
   go test -tags integration ./installer
   ```
