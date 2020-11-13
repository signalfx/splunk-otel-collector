# Build splunk-otel-collector rpm package

Build the splunk-otel-collector rpm package with [fpm](https://github.com/jordansissel/fpm).

To build the rpm package, run `make rpm-package` from the repo root directory. The rpm package will be written to
`dist/splunk-otel-collector-<version>.<arch>>.rpm`.

By default, `<arch>` is `amd64` and `<version>` is the latest git tag with `~post` appended, e.g. `1.2.3~post`.
To override these defaults, set the `ARCH` and `VERSION` environment variables, e.g.
`make rpm-package ARCH=arm64 VERSION=4.5.6`.

Run `pytest -m rpm internal/buildscripts/packaging/tests/package_test.py` to run basic installation tests with the built
package. See [README.md](../../tests/README.md) for how to install `pytest` and more details.
