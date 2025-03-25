#!/bin/bash

set -euo pipefail

mkdir -p ./govulncheck 2>/dev/null

# Get all package directories
ALL_PKG_DIRS=$(go list ./...)

# Initialize failure flag
FAILED=0

# Repository prefix to remove from package names
REPO_PREFIX=$(go list -m)

# Run govulncheck for each package
for pkg in $ALL_PKG_DIRS; do
  # Remove the repository prefix from the package name to keep the category names short
  # and replace slashes with underscores to make clear that the categories are not nested.
  OUTPUT_FILE="./govulncheck/$(echo "$pkg" | sed "s|^$REPO_PREFIX/||" | tr '/' '_').sarif"
  echo -e "\nRunning govulncheck for package $pkg"
  if ! govulncheck ${GOVULN_OPTS:-} "$pkg" > "$OUTPUT_FILE"; then
    echo "govulncheck failed for package $pkg, output saved to $OUTPUT_FILE"
    FAILED=1
  else
    echo "govulncheck succeeded for package $pkg, output saved to $OUTPUT_FILE"
  fi
done

if [ $FAILED -ne 0 ]; then
  echo -e "\ngovulncheck failed for one or more packages"
  exit 1
fi

echo -e "\ngovulncheck completed successfully for all packages"
