#!/bin/bash

# Copyright Splunk Inc.
#
# Helper script to build Windows MSI for local testing
# This script simplifies the build process for Puppet testing

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$( cd "$SCRIPT_DIR/../../" && pwd )"

# Default version for local testing
DEFAULT_VERSION="0.0.1-local"
VERSION="${1:-$DEFAULT_VERSION}"

echo "============================================"
echo "Building Windows MSI for Local Testing"
echo "============================================"
echo ""
echo "Version: $VERSION"
echo "Repository: $REPO_DIR"
echo ""

# Check prerequisites
echo "Checking prerequisites..."
echo ""

if [[ "$OSTYPE" != "msys" && "$OSTYPE" != "cygwin" && "$OSTYPE" != "win32" ]]; then
    echo "ERROR: This script must be run on Windows (Git Bash, WSL, or similar)"
    echo ""
    echo "You are currently on: $OSTYPE"
    echo ""
    echo "Please run this script on a Windows machine with:"
    echo "  - Git Bash"
    echo "  - WiX Toolset 3.14.0.8606"
    echo "  - Visual Studio or Build Tools for Visual Studio"
    exit 1
fi

# Check for required files
echo "Step 1/3: Checking for Windows binary..."
if [[ ! -f "$REPO_DIR/bin/otelcol_windows_amd64.exe" ]]; then
    echo "ERROR: Windows binary not found at: $REPO_DIR/bin/otelcol_windows_amd64.exe"
    echo ""
    echo "Please build the binary first:"
    echo "  make binaries-windows_amd64"
    exit 1
fi
echo "✓ Found Windows binary"
echo ""

echo "Step 2/3: Checking for agent bundle..."
if [[ ! -f "$REPO_DIR/dist/agent-bundle_windows_amd64.zip" ]]; then
    echo "ERROR: Agent bundle not found at: $REPO_DIR/dist/agent-bundle_windows_amd64.zip"
    echo ""
    echo "Please build the agent bundle first (must be done on Windows):"
    echo "  cd packaging/bundle/scripts/windows"
    echo "  ./make.ps1 bundle"
    exit 1
fi
echo "✓ Found agent bundle"
echo ""

echo "Step 3/3: Building MSI..."
echo ""

# Build the MSI
"$SCRIPT_DIR/build.sh" "$VERSION"

echo ""
echo "============================================"
echo "Build Complete!"
echo "============================================"
echo ""

# Find the built MSI
MSI_FILE=$(find "$REPO_DIR/dist" -name "splunk-otel-collector-*-amd64.msi" -type f | head -n 1)

if [[ -n "$MSI_FILE" ]]; then
    MSI_FILENAME=$(basename "$MSI_FILE")
    # Extract version from filename: splunk-otel-collector-VERSION-amd64.msi
    MSI_VERSION=$(echo "$MSI_FILENAME" | sed 's/splunk-otel-collector-\(.*\)-amd64.msi/\1/')
    
    echo "MSI created: $MSI_FILE"
    echo "MSI version: $MSI_VERSION"
    echo ""
    echo "To test with Puppet:"
    echo ""
    echo "1. Start the local MSI server (in one terminal):"
    echo "   python packaging/tests/deployments/puppet/local_windows_test.py --serve"
    echo ""
    echo "2. Run the puppet tests (in another terminal):"
    echo "   WIN_COLLECTOR_VERSION=$MSI_VERSION LOCAL_MSI_SERVER=http://localhost:8000 \\"
    echo "   pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default"
    echo ""
else
    echo "ERROR: MSI file not found in $REPO_DIR/dist"
    echo "Build may have failed. Check the output above for errors."
    exit 1
fi

