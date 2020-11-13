# Package and Installer Tests

## Setup

1. Install Docker, Python 3, [pip](https://pip.pypa.io/en/stable/installing/),
   and [virtualenv](https://virtualenv.pypa.io/en/latest/) on your workstation.
1. Install the required dependencies with pip in virtualenv on your workstation:
   ```
   virtualenv venv
   source venv/bin/activate
   pip install -r internal/buildscripts/packaging/tests/requirements.txt
   ```
1. Check [pytest.org](https://pytest.org) or run `pytest --help` to see the
   available pytest options.

## Running tests

1. To run the package tests, execute the following commands:
   ```
   source venv/bin/activate  # if not already in virtualenv
   pytest [PYTEST_OPTIONS] internal/buildscripts/packaging/tests/package_test.py
   ```
   The package tests require that the deb and rpm packages to be tested exist
   locally in `<repo_base_dir>/dist`.  See [here](../fpm/deb/README.md) and
   [here](../fpm/rpm/README.md) for how to build the deb and rpm packages.
1. To run the installer tests, execute the following commands:
   ```
   source venv/bin/activate  # if not already in virtualenv
   pytest [PYTEST_OPTIONS] internal/buildscripts/packaging/tests/installer_test.py
   ```
