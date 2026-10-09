# Puppet deployment tests

The deployment tests are Go tests in [`tests/puppet`](../../../../tests/puppet).
They cover default and custom Collector settings, auto-instrumentation with and
without systemd, migration from libsplunk to the otel injector, and Windows
service configuration.

On Linux, Docker must be available. From the `tests` directory:

```bash
DEPLOYMENT_TEST_DISTRO=debian-bookworm DEPLOYMENT_TEST_GROUP=base \
  go test -tags puppet_integration -v -timeout 90m -count 1 ./puppet
DEPLOYMENT_TEST_DISTRO=debian-bookworm DEPLOYMENT_TEST_GROUP=instrumentation \
  go test -tags puppet_integration -v -timeout 90m -count 1 ./puppet
```

Set `VERSION`, `AUTO_INSTRUMENTATION_VERSION`, and `PUPPET_RELEASE` as needed.
The GitHub workflow builds the packages and sets these values automatically.
On Windows, set `DEPLOYMENT_TEST_CASE=default` or `custom` and run the same
`go test` command. The Windows tests require Chocolatey and Puppet.
