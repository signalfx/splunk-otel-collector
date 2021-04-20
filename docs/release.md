# Release Process

## Prerequisites and Setup

1. Connect to the Splunk corporate network on your workstation.
1. You must have permissions to push tags to this repository.
1. You must be able to sign git commits/tags. Follow [this guide](
   https://docs.github.com/en/github/authenticating-to-github/signing-commits)
   to set it up.
1. Create a Github access token that can write to this repository by going to
   [Personal Access tokens](https://github.com/settings/tokens) on Github.
   Save the token as the following environment variable:
   - `GITHUB_TOKEN`
1. You must have access to the required service account tokens for Chaperone,
   Splunk staging repository, and Splunk Artifactory.  Check with a team member
   for details.  Save the tokens as the following environment variables:
   - `ARTIFACTORY_TOKEN`
   - `CHAPERONE_TOKEN`
   - `STAGING_TOKEN`
1. You must have access to the S3 bucket.  Add a profile called `prod` to your
   AWS CLI tool config that contains your IAM credentials to our production AWS
   account.  The default region does not matter because we only deal with S3
   and CloudFront, which are region-less. This is generally done by adding a
   section with the header `[prod]` in the file `~/.aws/credentials`.
1. Install Docker, Python 3, [pip](https://pip.pypa.io/en/stable/installing/),
   and [virtualenv](https://virtualenv.pypa.io/en/latest/) on your workstation.
1. Install the required dependencies with pip in virtualenv on your workstation:
   ```
   $ virtualenv venv
   $ source venv/bin/activate
   $ pip install -r internal/buildscripts/packaging/release/requirements.txt
   ```

## Steps

1. If necessary, update the OpenTelemetry Core and Contrib dependency versions
   in [go.mod](../go.mod) and run `go mod tidy`.  Create a PR with the changes
   and ensure that the build and tests are successful.  Wait for the PR to be
   approved and merged, and ensure that the `main` branch build and tests are
   successful.
1. Create and push the tag with the appropriate version:
   ```
   $ make add-tag TAG=v1.2.3
   $ git push --tags origin  # assuming "origin" is the upstream repository and not your fork
   ```
1. Ensure that the build and tests for the tag are successful.
1. Ensure that the `quay.io/signalfx/splunk-otel-collector:<VERSION>` image
   was built and pushed.
1. Ensure that the Github draft release was created and that the unsigned
   packages were published to [Github Releases](
   https://github.com/signalfx/splunk-otel-collector/releases/).
1. Run the following script in virtualenv to download the packages from the
   Github draft release, sign them, push the signed packages to
   Artifactory, S3, and back to the Github draft release.
   ```
   $ source venv/bin/activate  # if not already in virtualenv
   $ ./internal/buildscripts/packaging/release/sign_release.py --stage release
   ```
   This may take 10+ minutes to complete.  Upon completion, the Github draft
   release will be published.  Run the script with `--help` for more details
   and to see all available options.
