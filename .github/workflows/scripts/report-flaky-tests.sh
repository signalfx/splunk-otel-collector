#!/bin/bash

# Reports flaky unit tests as GitHub issues, one open issue per flaky test.
#
# Input is the directory holding the unit-test-results-<os> artifacts produced by
# reusable-unit-test.yml. Tests there run through gotestsum with re-runs enabled
# (see TEST_RERUN_FAILS in Makefile.Common), so a test that both failed and passed
# within the same run is flaky, as opposed to a test that only ever failed, which
# is consistently broken and is not reported here.

set -euo pipefail

RESULTS_DIR="${1:?usage: $0 <test-results-dir>}"
# Label used to find the issues created by previous runs, do not change it without
# also relabeling the existing issues.
LABEL="${FLAKY_TEST_LABEL:-flaky test}"
# Guards against flooding the tracker if something goes wrong upstream.
MAX_NEW_ISSUES="${MAX_NEW_ISSUES:-10}"
REPO="${GITHUB_REPOSITORY:?}"
RUN_URL="${GITHUB_SERVER_URL:?}/${REPO}/actions/runs/${GITHUB_RUN_ID:?}"
COMMIT="${COMMIT_SHA:-${GITHUB_SHA:?}}"
# Set on pull request builds only.
PR_NUMBER="${PR_NUMBER:-}"

# Emits "<os> <package> <test> <action>" for every test result in every artifact.
test_results() {
  local artifact os results
  for artifact in "${RESULTS_DIR}"/*/; do
    [[ -d "${artifact}" ]] || continue
    os="$(basename "${artifact}")"
    os="${os#unit-test-results-}"
    while IFS= read -r results; do
      # A truncated file, from an interrupted test run, is reported and skipped
      # instead of failing the whole report.
      jq -rn --arg os "${os}" 'inputs
        | select(.Test != null and (.Action == "pass" or .Action == "fail"))
        | [$os, .Package, .Test, .Action] | @tsv' "${results}" ||
        echo "Ignored ${results}, it could not be parsed." >&2
    done < <(find "${artifact}" -name '*.json')
  done
}

# Emits "<package> <test> <os>[, <os>...]" for every test that both failed and
# passed on the same OS.
flaky_tests() {
  test_results | sort -u | awk -F'\t' '
    { results[$1 FS $2 FS $3 FS $4] = 1; tests[$1 FS $2 FS $3] = 1 }
    END {
      for (test in tests) {
        if ((test FS "pass") in results && (test FS "fail") in results) {
          print test
        }
      }
    }' | awk -F'\t' '
    # A failing sub-test fails its parents too, only the sub-test is reported.
    { flaky[$0] = 1
      if (index($3, "/") > 0) {
        count = split($3, name, "/")
        parent = name[1]
        for (i = 1; i < count; i++) {
          parents[$1 FS $2 FS parent] = 1
          parent = parent "/" name[i + 1]
        }
      }
    }
    END { for (test in flaky) if (!(test in parents)) print test }' | sort |
    awk -F'\t' '
    { key = $2 FS $3; oses[key] = (key in oses) ? oses[key] ", " $1 : $1 }
    END { for (key in oses) print key FS oses[key] }' | sort
}

if [[ ! -d "${RESULTS_DIR}" ]]; then
  echo "No test results in ${RESULTS_DIR}, nothing to report."
  exit 0
fi

flaky="$(mktemp)"
flaky_tests > "${flaky}"
if [[ ! -s "${flaky}" ]]; then
  echo "No flaky tests found."
  exit 0
fi
echo "Flaky tests found:"
cat "${flaky}"

gh label create "${LABEL}" --repo "${REPO}" --color "fbca04" \
  --description "Test that fails intermittently" >/dev/null 2>&1 ||
  echo "Label \"${LABEL}\" already exists."

# Issues closed by a fix are not reopened, a new one is created if the test flakes
# again, so only open issues are matched.
open_issues="$(mktemp)"
gh issue list --repo "${REPO}" --state open --label "${LABEL}" --limit 500 \
  --json number,title > "${open_issues}"

created=0
while IFS=$'\t' read -r package test oses; do
  title="Flaky test: ${test} in ${package}"
  occurrence="$(
    printf -- '- Failed and passed on re-run on: %s\n- Commit: %s\n- Build: %s\n' \
      "${oses}" "${COMMIT}" "${RUN_URL}"
    if [[ -n "${PR_NUMBER}" ]]; then
      printf -- '- Pull request: #%s\n' "${PR_NUMBER}"
    fi
  )"
  number="$(jq -r --arg title "${title}" \
    'map(select(.title == $title)) | .[0].number // empty' "${open_issues}")"

  if [[ -n "${number}" ]]; then
    echo "Adding occurrence to issue #${number}: ${title}"
    gh issue comment "${number}" --repo "${REPO}" --body "${occurrence}"
  elif (( created >= MAX_NEW_ISSUES )); then
    echo "Reached MAX_NEW_ISSUES=${MAX_NEW_ISSUES}, skipping: ${title}"
  else
    echo "Creating issue: ${title}"
    gh issue create --repo "${REPO}" --title "${title}" --label "${LABEL}" \
      --body "$(printf -- '`%s` in `%s` failed and then passed when re-run, which means it is flaky.\n\n%s\nFurther occurrences are added as comments to this issue.\n' \
        "${test}" "${package}" "${occurrence}")"
    created=$((created + 1))
  fi
done < "${flaky}"
