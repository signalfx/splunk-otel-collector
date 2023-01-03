#!/bin/bash
echo "starting tests!"
set -euxo pipefail
# Tests that github actions env vars are set as expected.  Likely to fail on first few pushes, as we will be validating
# an error case to start with.

echo "GITHUB_WORKSPACE=$GITHUB_WORKSPACE"
echo "GITHUB_REF_NAME=$GITHUB_REF_NAME"
echo "REF_TYPE=$REF_TYPE"
echo "exiting test.."