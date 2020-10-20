# Release Process

## Prerequisites and Setup

1. Connect to the Splunk corporate network on your workstation.
1. You must have permissions to push tags to this repository.
1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/github/authenticating-to-github/signing-commits)
   to setup it up.
1. Access the required service account tokens for Chaperone, Splunk staging
   repository, and Splunk Artifactory.  Check with a team member for details.
1. Install Docker, Python 3, [pip](https://pip.pypa.io/en/stable/installing/),
   and [virtualenv](https://virtualenv.pypa.io/en/latest/) on your workstation.
1. Install the required dependencies with pip in virtualenv on your workstation:
   ```
   virtualenv venv
   source venv/bin/activate
   pip install -r internal/buildscripts/packaging/release/requirements.txt
   ```

## Steps

1. If necessary, update the OpenTelemetry Core and Contrib dependency versions
   in [go.mod](../go.mod) and run `go mod tidy`.  Create a PR with the changes
   and ensure that the build and tests are successful.  Wait for the PR to be
   approved and merged, and ensure that the `main` branch build and tests are
   successful.
1. Create and push the tag with the appropriate version:
   ```
   make add-tag TAG=v1.2.3
   git push --tags origin  # assuming "origin" is the upstream repository and not your fork
   ```
1. Ensure that the build and tests for the tag are successful.
1. Ensure that the release was created and the packages were published to
   [Github Releases](https://github.com/signalfx/splunk-otel-collector/releases/)
1. Run the following script in virtualenv with the respective service account
   tokens to sign the packages from github and release them to Artifactory.
   `STAGE` should be `test`, `beta`, or `release` (default).
   ```
   source venv/bin/activate  # if not already in virtualenv
   ./internal/buildscripts/packaging/release/sign_release.py --stage=STAGE --artifactory-token=ARTIFACTORY_TOKEN --chaperone-token=CHAPERONE_TOKEN --staging-token=STAGING_TOKEN
   ```
   This may 10+ minutes to complete.

   **Hint:** You can set the `ARTIFACTORY_TOKEN`, `CHAPERONE_TOKEN`, and
   `STAGING_TOKEN` environment variables instead of passing the tokens via
   the command line options.