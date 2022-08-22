# Build splunk-otel-collector tar package

Build the splunk-otel-collector tar package with [fpm](https://github.com/jordansissel/fpm).

To build the tar package, run `make tar-package` from the repo root directory. The tar package will be written to
`dist/splunk-otel-collector_<version>_<arch>.tar.gz`.

By default, `<arch>` is `amd64` and `<version>` is the latest git tag with `-post` appended, e.g. `1.2.3-post`.
To override these defaults, set the `ARCH` and `VERSION` environment variables, e.g.
`make tar-package ARCH=arm64 VERSION=4.5.6`.

Run `pytest -m tar internal/buildscripts/packaging/tests/package_test.py` to run basic installation tests with the built
package. See [README.md](../../tests/README.md) for how to install `pytest` and more details.
