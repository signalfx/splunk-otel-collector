# Release Process

## Prerequisites and Setup

1. Connect to the Splunk corporate network on your workstation.
1. You must have permissions to push tags to this repository.
1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/github/authenticating-to-github/signing-commits)
   to set it up.
1. You must have access to the `o11y-gdi/splunk-otel-collector-releaser` gitlab
   repo and CI/CD pipeline.
1. You must have an AWS access key that has permissions to push to the SignalFx
   S3 bucket, i.e. `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.
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
   (may take up to 30 minutes to complete).
1. Ensure that the `quay.io/signalfx/splunk-otel-collector:<VERSION>` image
   was built and pushed.
1. Check [Github Releases](
   https://github.com/signalfx/splunk-otel-collector/releases/) and ensure that
   the release was created for the tag and all assets were uploaded.  Update
   the release notes for the Github Release with the changes from
   [CHANGELOG.md](../CHANGELOG.md).
1. Download the MSI (`splunk-otel-collector-<VERSION>-amd64.msi`) from the
   Github Release to your workstation.
1. Run the following script in virtualenv to push the signed MSI and installer
   scripts to S3 (replace `PATH_TO_MSI` with the path to the signed MSI file
   downloaded from the previous step, and `AWS_ACCESS_KEY_ID` and
   `AWS_SECRET_ACCESS_KEY` with your personal AWS access key).
   ```
   $ source venv/bin/activate  # if not already in virtualenv
   $ ./internal/buildscripts/packaging/release/sign_release.py --stage=release --path=PATH_TO_MSI --installers --no-sign-msi --aws-key-id=AWS_ACCESS_KEY_ID --aws-key=AWS_SECRET_ACCESS_KEY
   ```
