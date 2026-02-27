#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <path-to-packages>"
    exit 1
fi

PACKAGES_PATH="$(realpath "$1")"

if [ ! -d "$PACKAGES_PATH" ]; then
    echo "Error: '$PACKAGES_PATH' is not a directory"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
IMAGE_NAME="splunk-appinspect"

echo "Building Docker image from $SCRIPT_DIR/Dockerfile.splunk-appinspect..."
docker build -t "$IMAGE_NAME" -f "$SCRIPT_DIR/Dockerfile.splunk-appinspect" "$SCRIPT_DIR"

echo "Running appinspect on packages in $PACKAGES_PATH..."
docker run --rm -v "$PACKAGES_PATH:/packages:ro" "$IMAGE_NAME"
