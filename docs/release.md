# Release Process

## Prerequisites and Setup

1. Connect to the Splunk corporate network on your workstation.
1. You must have permissions to push tags to this repository.
1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)
   to set it up.
1. You must have access to the `o11y-gdi/splunk-otel-collector-releaser` gitlab
   repo and CI/CD pipeline.
1. You must have prod access and the `splunkcloud_account_power` role in `us0`
   to push to the SignalFx S3 bucket.
1. Install Python 3, [pip](https://pip.pypa.io/en/stable/installing/),
   and [virtualenv](https://virtualenv.pypa.io/en/latest/) on your workstation.
1. Install the required dependencies with pip in virtualenv on your workstation:
   ```
   $ virtualenv venv
   $ source venv/bin/activate
   $ pip install -r internal/buildscripts/packaging/release/requirements.txt
   ```

## Steps

1. If necessary, update the OpenTelemetry Core and Contrib dependency versions
   in [go.mod](../go.mod) and run `go mod tidy`.
1. If necessary, update [smart-agent-release.txt](
   ../internal/buildscripts/packaging/smart-agent-release.txt) for the latest
   applicable [Smart Agent release](
   https://github.com/signalfx/signalfx-agent/releases).
1. Update [CHANGELOG.md](../CHANGELOG.md) with the changes for the release.
   In order for the Github release notes to be added correctly, ensure that the
   new version has the `## <TAG>` heading.
1. Create a PR with the changes and ensure that the build and tests are
   successful.  Wait for the PR to be approved and merged, and ensure that the
   `main` branch build and tests are successful.
1. Create and push the tag with the appropriate version:
   ```
   $ make add-tag TAG=v1.2.3
   $ git push --tags origin  # assuming "origin" is the upstream repository and not your fork
   ```
1. Wait for the gitlab repo to be synced with the new tag (may take up to 30
   minutes; if you have permissions, you can trigger the sync immediately from
   repo settings in gitlab).  The CI/CD pipeline will then trigger
   automatically for the new tag.
1. Ensure that the build and release jobs in gitlab for the tag are successful
   (may take over 30 minutes to complete).
1. Ensure that the `quay.io/signalfx/splunk-otel-collector:<VERSION>` image
   was built and pushed.
1. Ensure that the `quay.io/signalfx/splunk-otel-collector-windows:<VERSION>`
   image was built and pushed.
1. Check [Github Releases](
   https://github.com/signalfx/splunk-otel-collector/releases/) and ensure that
   the release was created for the tag, all assets were uploaded, and the
   release notes match the changes for the version in [CHANGELOG.md](
   ../CHANGELOG.md).
1. Download the MSI (`splunk-otel-collector-<VERSION>-amd64.msi`) from the
   Github Release to your workstation.
1. Request prod access via slack and the `splunkcloud_account_power` role with
   `okta-aws-setup us0`.
1. Run the following script in virtualenv to push the signed MSI and installer
   scripts to S3 (replace `PATH_TO_MSI` with the path to the signed MSI file
   downloaded from the previous step).
   ```
   $ source venv/bin/activate  # if not already in virtualenv
   $ ./internal/buildscripts/packaging/release/sign_release.py --stage=release --path=PATH_TO_MSI --installers --no-sign-msi
   ```
