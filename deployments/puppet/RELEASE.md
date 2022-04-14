# Release Process

## Prerequisites

1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)
   to set it up.
1. You must have access to the `o11y-gdi/splunk-otel-collector-releaser` gitlab
   repo and CI/CD pipeline.

## Steps

1. Ensure that the version in [metadata.json](./metadata.json) has been updated
   to an appropriate semver.  If necessary, create a PR with this change and
   wait for it to be approved and merged.
1. Check [Github Actions](
   https://github.com/signalfx/splunk-otel-collector/actions/workflows/puppet.yml)
   and ensure that the workflow completed successfully.  A new
   `puppet-v<VERSION>` tag should be pushed, where `VERSION` is the version
   from [metadata.json](./metadata.json).
1. Wait for the gitlab repo to be synced with the new tag (may take up to 30
   minutes; if you have permissions, you can trigger the sync immediately from
   repo settings in gitlab).  The CI/CD pipeline will then trigger
   automatically for the new tag.
1. Check the gitlab CI/CD pipeline and ensure that the `puppet-release` job
   completes successfully.
1. Check [https://forge.puppet.com/modules/signalfx/splunk_otel_collector](
   https://forge.puppet.com/modules/signalfx/splunk_otel_collector) and ensure
   that the release was published successfully.
