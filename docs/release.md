# Release Process

## Prerequisites and Setup

1. Connect to the Splunk corporate network on your workstation.
1. You must have permissions to push tags to this repository.
1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/github/authenticating-to-github/signing-commits)
   to set it up.
1. You must have access to the `o11y-gdi/splunk-otel-collector-releaser` gitlab
   repo and CI/CD pipeline.
1. You must have prod access and the `splunkcloud_account_power` role in `us0`
   to push to the SignalFx S3 bucket.
1. Install Golang (same version as the `go` directive in [go.mod](../go.mod)).
1. Install Python 3, [pip](https://pip.pypa.io/en/stable/installing/),
   and [virtualenv](https://virtualenv.pypa.io/en/latest/) on your workstation.
1. Install the required dependencies with pip in virtualenv on your workstation:
   ```
   $ virtualenv venv
   $ source venv/bin/activate
   $ pip install -r internal/buildscripts/packaging/release/requirements.txt
   ```
1. Clone cargo repository for local access to the okta-aws-setup script.

## Steps

1. If necessary, update the OpenTelemetry Core and Contrib dependency versions
   in [go.mod](../go.mod) and [tests/go.mod](../tests/go.mod), and run
   `make tidy`.  Alternatively, you can try using the [update-deps](
   ../internal/buildscripts/update-deps) script (replace `<VERSION>` with
   either `latest` or the desired Core/Contrib tag):
   ```
   $ OTEL_VERSION=<VERSION> ./internal/buildscripts/update-deps
   ```
   Check the changes to ensure that everything was updated as intended.
   **Note:** The script will try to update each Core/Contrib dependency in
   `go.mod` one at a time, and may fail due to breaking upstream changes.
1. If necessary, update [smart-agent-release.txt](
   ../internal/buildscripts/packaging/smart-agent-release.txt) for the latest
   applicable [Smart Agent release](
   https://github.com/signalfx/signalfx-agent/releases).
1. If the Smart Agent from the previous step was updated, update the
   `github.com/signalfx/signalfx-agent` and
   `github.com/signalfx/signalfx-agent/pkg/apm` dependencies in [go.mod](
   ../go.mod) to the commit hash for the Smart Agent release tag, and run
   `make tidy`.
1. If necessary, update [java-agent-release.txt](
   ../instrumentation/packaging/java-agent-release.txt) for the latest
   applicable [Java Agent release](
   https://github.com/signalfx/splunk-otel-java/releases).
1. Update [CHANGELOG.md](../CHANGELOG.md) with the changes for the release.
   - This requires going through PRs merged since the last release, as the
   CHANGELOG may not be properly updated.
   - In order for the Github release notes to be added correctly, ensure that
   the new version has the `## <TAG>` heading.
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
   - Make sure to check the pipeline for the new tag, not the commit on the
     `main` branch.
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
1. Request the `signalfx/splunkcloud_account_power` role by running
   `okta-aws-setup us0` locally on your workstation.
1. Run the following script in virtualenv to push the signed MSI and installer
   scripts to S3 (replace `PATH_TO_MSI` with the path to the signed MSI file
   downloaded from the previous step).
   ```
   $ source venv/bin/activate  # if not already in virtualenv
   $ ./internal/buildscripts/packaging/release/sign_release.py --stage=release --path=PATH_TO_MSI --installers --no-sign-msi
   ```
   - The script may ask to overwrite existing files. Opt `y` if prompted. You should not
     be prompted to overwrite a `msi` file unless you're purposefully trying to
     replace the `msi` file for an existing release.
   - Operation is successful if no exception or traceback is shown in script output.
   - You can confirm success by downloading the MSI from this link:
   ```
   https://dl.signalfx.com/splunk-otel-collector/msi/release/splunk-otel-collector-VERSION-amd64.msi
   VERSION - Formatted as 1.2.3
   ```

## Ansible/Chef/Puppet Release Steps

These are ad-hoc releases to make when specific modules are changed, independent
from the collector releases. Authors for changes to these modules are responsible
to update these releases.

1. Open a PR in a non-forked branch with the updated version and changelog
   for the module:
   - Ansible: [galaxy.yml](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/ansible/galaxy.yml)
   - Chef: [metadata.rb](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/chef/metadata.rb)
   - Puppet: [metadata.json](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/puppet/metadata.json)
1. After the PR is merged, a new tag based on the new version will be created
   and pushed by the corresponding github workflow:
   - [Ansible](https://github.com/signalfx/splunk-otel-collector/actions/workflows/ansible.yml)
   - [Chef](https://github.com/signalfx/splunk-otel-collector/actions/workflows/chef.yml)
   - [Puppet](https://github.com/signalfx/splunk-otel-collector/actions/workflows/puppet.yml)
1. The corresponding gitlab release pipeline will be triggered for the new tag
   (may take up to 30 minutes to sync with github) to build and publish the
   new module version:
   - [Ansible Galaxy](https://galaxy.ansible.com/signalfx/splunk_otel_collector)
   - [Chef Supermarket](https://supermarket.chef.io/cookbooks/splunk_otel_collector)
   - [Puppet Forge](https://forge.puppet.com/modules/signalfx/splunk_otel_collector)
