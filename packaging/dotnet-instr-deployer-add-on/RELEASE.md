# Release Process

## Prerequisites

1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/github/authenticating-to-github/signing-commits)
   to set it up.
1. You must have access to the `o11y-gdi/splunk-otel-collector-releaser` gitlab
   repo and CI/CD pipeline.

## Steps

1. Ensure that the version in [assets/default/app.conf](./assets/default/app.conf)
   has been updated to an appropriate semver. If necessary, create a PR with this
   change and wait for it to be approved and merged.
1. Check [Github Actions](
   https://github.com/signalfx/splunk-otel-collector/actions/workflows/dotnet-instr-deployer-add-on.yml)
   and ensure that the workflow completed successfully.  A new
   `dotnet-instrumentation-deployer-ta-v<VERSION>` tag should be pushed,
   where `VERSION` is the version from [assets/default/app.conf](./assets/default/app.conf).
1. Wait for the gitlab repo to be synced with the new tag (may take up to 30
   minutes; if you have permissions, you can trigger the sync immediately from
   repo settings in gitlab).  The CI/CD pipeline will then trigger
   automatically for the new tag.
1. Check the gitlab CI/CD pipeline and ensure that the `dotnet-instrumentation-deployer-release`
   job completes successfully.
1. The release is created as `prerelease` and `draft` on GitHub.  You can
   publish it by clicking on the `Publish release` button.  The release notes are
   automatically generated from the changelog in [CHANGELOG.md](./CHANGELOG.md).
