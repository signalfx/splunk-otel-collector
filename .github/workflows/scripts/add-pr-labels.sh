#!/usr/bin/env bash
#
# Copyright The OpenTelemetry Authors
# SPDX-License-Identifier: Apache-2.0
#
# Adds relevant labels to a PR.
# The original purpose of this script is to accommodate workflows to
# be triggered when either relevant files are modified, or a label
# is added. GitHub's workflow logic requires all parameters to match
# when triggering runs, so an OR condition doesn't work.
# Note that since this script is considered a requirement for PRs,
# it should never fail.

set -euo pipefail

main () {
    LABELS=""
    VULN_SCAN_LABEL="Run vulnerabilities scan"

    JSON=$(gh pr view "${PR}" --json "files" | tr -dc '[:print:]' | sed -E 's/\\[a-z]//g')
    FILES=$(echo -n "${JSON}"| jq -r '.files[].path')

    for FILE in ${FILES}; do
        case "${FILE}" in
            ".github/workflows/vuln-scans.yml"|".grype.yaml"|".trivyignore"|".snyk")
                # Only add label once
                if [[ "${LABELS}" != *"${VULN_SCAN_LABEL}"* ]]; then
                  if [[ -n "${LABELS}" ]]; then
                      LABELS+=","
                  fi
                  LABELS+=${VULN_SCAN_LABEL}
                fi
                ;;
            *)
                echo "No match found"
                ;;
        esac
    done

    if [[ -n "${LABELS}" ]]; then
        echo "Adding labels: ${LABELS}"
         gh pr edit "${PR}" --add-label "${LABELS}" || echo "Failed to add labels"
    else
        echo "No labels found"
    fi
}

# We don't want this workflow to ever fail and block a PR,
# so ensure all errors are caught.
main || echo "Failed to run $0"
