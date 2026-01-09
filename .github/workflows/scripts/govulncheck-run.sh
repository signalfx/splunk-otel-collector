#!/bin/bash

set -eo pipefail

mkdir -p ./govulncheck 2>/dev/null

# Get all package directories
ALL_PKG_DIRS=$(go list ./...)

# Initialize failure flag
FAILED=0

# Repository prefix to remove from package names
REPO_PREFIX=$(go list -m)

# Use a bash regex to extract the the value of the --format flag
# from the GOVULN_OPTS environment variable
if [[ "$GOVULN_OPTS" =~ .*--format[[:space:]]+([a-z]+).* ]]; then
  FORMAT=${BASH_REMATCH[1]}
fi

# Run govulncheck for each package
for pkg in $ALL_PKG_DIRS; do
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
done

if [ $FAILED -ne 0 ]; then
  echo -e "\n**** govulncheck failed for one or more packages"
  exit 1
fi

echo -e "\n**** govulncheck completed successfully for all packages"
