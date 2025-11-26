#!/usr/bin/env python3
"""
Local Windows Testing Helper for Puppet Tests

This script helps you test puppet deployment with locally built Windows MSI artifacts.

Usage:
    1. Build Windows artifacts (run from repository root):
       - On Windows: make binaries-windows_amd64
       - Build agent bundle: packaging/bundle/scripts/windows/make.ps1 bundle
       - Build MSI: packaging/msi/build.sh [VERSION]
    
    2. Run this script to start a local HTTP server:
       python local_windows_test.py --serve
    
    3. In another terminal, run the puppet tests with the local server:
       WIN_COLLECTOR_VERSION=<your_version> LOCAL_MSI_SERVER=http://localhost:8000 \
       pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default
"""

import argparse
import http.server
import os
import socketserver
import sys
from pathlib import Path

# Add the tests directory to Python path
REPO_DIR = Path(__file__).parent.parent.parent.parent.parent.resolve()
sys.path.insert(0, str(REPO_DIR))

DIST_DIR = REPO_DIR / "dist"
DEFAULT_PORT = 8000


class MSIHandler(http.server.SimpleHTTPRequestHandler):
    """Custom handler to serve MSI files from dist directory"""
    
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(DIST_DIR), **kwargs)
    
    def log_message(self, format, *args):
        """Log with more detail"""
        sys.stdout.write(f"[MSI Server] {self.address_string()} - {format % args}\n")
        sys.stdout.flush()


def find_msi_files():
    """Find all MSI files in the dist directory"""
    if not DIST_DIR.exists():
        return []
    return list(DIST_DIR.glob("*.msi"))


def extract_version_from_msi(msi_path):
    """Extract version from MSI filename"""
    # Expected format: splunk-otel-collector-X.Y.Z-amd64.msi
    filename = msi_path.name
    if filename.startswith("splunk-otel-collector-") and filename.endswith("-amd64.msi"):
        version = filename.replace("splunk-otel-collector-", "").replace("-amd64.msi", "")
        return version
    return None


def serve_msi(port=DEFAULT_PORT):
    """Start HTTP server to serve MSI files"""
    msi_files = find_msi_files()
    
    if not msi_files:
        print(f"ERROR: No MSI files found in {DIST_DIR}")
        print("\nPlease build the MSI first:")
        print("  1. On Windows, run: make binaries-windows_amd64")
        print("  2. Build agent bundle: packaging/bundle/scripts/windows/make.ps1 bundle")
        print("  3. Build MSI: packaging/msi/build.sh")
        sys.exit(1)
    
    print(f"Found {len(msi_files)} MSI file(s) in {DIST_DIR}:")
    for msi in msi_files:
        version = extract_version_from_msi(msi)
        print(f"  - {msi.name} (version: {version})")
    
    print(f"\nStarting HTTP server on port {port}...")
    print(f"Serving files from: {DIST_DIR}")
    print(f"\nTo run puppet tests, use:")
    
    for msi in msi_files:
        version = extract_version_from_msi(msi)
        if version:
            print(f"\n  WIN_COLLECTOR_VERSION={version} LOCAL_MSI_SERVER=http://localhost:{port} \\")
            print(f"  pytest -v packaging/tests/deployments/puppet/puppet_test.py::test_win_puppet_default")
    
    print(f"\nPress Ctrl+C to stop the server\n")
    print("=" * 70)
    
    with socketserver.TCPServer(("", port), MSIHandler) as httpd:
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n\nShutting down server...")
            httpd.shutdown()


def check_artifacts():
    """Check if all required artifacts exist"""
    print("Checking for required Windows artifacts...")
    print(f"Distribution directory: {DIST_DIR}\n")
    
    required_files = [
        ("Windows Binary", "bin/otelcol_windows_amd64.exe"),
        ("Agent Bundle", "dist/agent-bundle_windows_amd64.zip"),
    ]
    
    all_found = True
    for name, rel_path in required_files:
        full_path = REPO_DIR / rel_path
        exists = full_path.exists()
        status = "✓ Found" if exists else "✗ Missing"
        print(f"{status}: {name} ({rel_path})")
        if not exists:
            all_found = False
    
    msi_files = find_msi_files()
    if msi_files:
        print(f"✓ Found: MSI files ({len(msi_files)} file(s))")
        for msi in msi_files:
            version = extract_version_from_msi(msi)
            print(f"    - {msi.name} (version: {version})")
    else:
        print("✗ Missing: MSI file in dist/")
        all_found = False
    
    print()
    
    if all_found:
        print("✓ All required artifacts found!")
        print("\nYou can now run:")
        print("  python local_windows_test.py --serve")
        return True
    else:
        print("✗ Some artifacts are missing. Please build them first.")
        print("\nBuild instructions:")
        print("  1. Build Windows binary:")
        print("     make binaries-windows_amd64")
        print("  2. Build agent bundle (on Windows):")
        print("     packaging/bundle/scripts/windows/make.ps1 bundle")
        print("  3. Build MSI (on Windows with WiX Toolset):")
        print("     packaging/msi/build.sh [VERSION]")
        return False


def main():
    parser = argparse.ArgumentParser(
        description="Helper script for testing Puppet deployment with local Windows MSI"
    )
    parser.add_argument(
        "--serve",
        action="store_true",
        help="Start HTTP server to serve MSI files"
    )
    parser.add_argument(
        "--port",
        type=int,
        default=DEFAULT_PORT,
        help=f"Port for HTTP server (default: {DEFAULT_PORT})"
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Check if all required artifacts exist"
    )
    
    args = parser.parse_args()
    
    if args.serve:
        serve_msi(args.port)
    elif args.check:
        check_artifacts()
    else:
        parser.print_help()
        print("\n" + "=" * 70)
        check_artifacts()


if __name__ == "__main__":
    main()

