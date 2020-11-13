# Build splunk-otel-collector deb package

Build the splunk-otel-collector deb package with [fpm](https://github.com/jordansissel/fpm).

To build the deb package, run `make deb-package` from the repo root directory. The deb package will be written to
`dist/splunk-otel-collector_<version>_<arch>.deb`.

By default, `<arch>` is `amd64` and `<version>` is the latest git tag with `-post` appended, e.g. `1.2.3-post`.
To override these defaults, set the `ARCH` and `VERSION` environment variables, e.g.
`make deb-package ARCH=arm64 VERSION=4.5.6`.

Run `pytest -m deb internal/buildscripts/packaging/tests/package_test.py` to run basic installation tests with the built
package. See [README.md](../../tests/README.md) for how to install `pytest` and more details.
