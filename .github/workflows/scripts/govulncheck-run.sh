#!/bin/bash

set -eo pipefail

mkdir -p ./govulncheck 2>/dev/null

while [[ $# -gt 0 ]]; do
  case "$1" in
    -p|--package)
      pkg="$2"
      shift 2 # Move past the flag and its value
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

# Initialize failure flag
FAILED=0

# Repository prefix to remove from package names
REPO_PREFIX=$(go list -m)

# Use a bash regex to extract the value of the --format flag
# from the GOVULN_OPTS environment variable
if [[ "$GOVULN_OPTS" =~ .*--format[[:space:]]+([a-z]+).* ]]; then
  FORMAT=${BASH_REMATCH[1]}
fi

echo -e "\n**** Running govulncheck for package $pkg"
set +e
if [[ -z $FORMAT ]]; then
  govulncheck ${GOVULN_OPTS} "$pkg"
else
  # Remove the repository prefix from the package name to keep the category names short
  # and replace slashes with underscores to make clear that the categories are not nested.
  OUTPUT_FILE="./govulncheck/$(echo "$pkg" | sed "s|^$REPO_PREFIX/||" | tr '/' '_').$FORMAT"
  govulncheck ${GOVULN_OPTS} "$pkg" > "$OUTPUT_FILE"
fi
if [ $? -eq 0 ]; then
  echo -e "\n**** govulncheck succeeded for package $pkg"
else
  echo -e "\n**** govulncheck failed for package $pkg"
  FAILED=1
fi
set -e

if [ $FAILED -ne 0 ]; then
  echo -e "\n**** govulncheck failed for one or more packages"
  exit 1
fi

echo -e "\n**** govulncheck completed successfully for all packages"
