#!/usr/bin/env bash
# Checks that all external `uses:` references in .github/ YAML files are pinned
# to a full 40-character commit SHA.
#
# Exits 0 if all references are pinned, 1 otherwise.
#
# When running inside GitHub Actions, emits ::error:: annotations so each
# violation is highlighted inline in the PR diff.

set -euo pipefail

SEARCH_DIR="${1:-.github}"

# Find every `uses:` directive that is:
#   - not a comment: grep -rn output is "file:line:content"; exclude lines
#     where the content (after "file:linenum:") starts with optional whitespace
#     then "#". This covers both "  # uses: ..." and "  - # uses: ..." forms.
#   - not a local reference (uses: ./.github/...)
#   - not already pinned to a 40-char hex SHA
#
# NOTE: do NOT anchor with ^\s*uses: — step-level uses directives appear as
#   "      - uses: owner/repo@ref" and the leading "- " breaks that pattern.
UNPINNED_FILE=$(mktemp)
grep -rn "uses:" "$SEARCH_DIR" \
  --include="*.yml" \
  --include="*.yaml" \
| grep -v ':[0-9]*:[[:space:]]*#' \
| grep "@" \
| grep -v 'uses:[[:space:]]*\.' \
| grep -v "@[0-9a-f]\{40\}" \
> "$UNPINNED_FILE" || true

if [[ ! -s "$UNPINNED_FILE" ]]; then
  rm -f "$UNPINNED_FILE"
  echo "All external actions are SHA-pinned."
  exit 0
fi

COUNT=$(wc -l < "$UNPINNED_FILE")
echo "Found ${COUNT} unpinned action reference(s):"
echo ""

while IFS= read -r line; do
  # grep -rn output format:  path/to/file.yml:42:    uses: owner/repo@vtag
  file="${line%%:*}"
  rest="${line#*:}"
  lineno="${rest%%:*}"
  content="${rest#*:}"
  ref="$(echo "$content" | sed 's/.*uses:[[:space:]]*//')"

  echo "  $file:$lineno"
  echo "    $content"
  echo ""

  # Emit a workflow annotation when running inside GitHub Actions
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    echo "::error file=${file},line=${lineno}::Unpinned action: ${ref} — pin to a full 40-char commit SHA, e.g. uses: ${ref%%@*}@<sha> # ${ref##*@}"
  fi
done < "$UNPINNED_FILE"
rm -f "$UNPINNED_FILE"

echo "Pin each reference to a full 40-character commit SHA, e.g.:"
echo "  uses: actions/checkout@93cb6efe18208431cddfb8368fd83d5badbf9bfd # v5"

exit 1
