# Making a new release of the TA

## Update versions for splunk otel collector && TA
1. run `make update-deps`
2. Verify `git diff main -- Makefile Splunk_TA_otel/default/app.conf` shows expected changes


## Cut Release
- run `make update-and-release` to automatically upgrade to the latest splunk otel collector release and bump the minor release.
    - For more control, you can directly specify the environment variables `SPLUNK_OTEL_VERSION` and `TA_VERSION`
    - Example: `SPLUNK_OTEL_VERSION=0.0.0 TA_VERSION=0.0.1 make -e update-and-release` (replacing the version values as desired)
- upload `out/distribution/Splunk_TA_otel.tgz` artifact from `build-all-platforms` job of release branch to splunkbase

If you have issues with pushing the branch due to duplicate refs or anything, `git tag -d TAG_NAME` can be useful

## Market release
1. Inform docs team of release (or DIY)
2. Include link to our [latest release notes](https://github.com/signalfx/splunk-otel-collector/releases/)
3. Optionally include link to [latest upstream release notes](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases)
