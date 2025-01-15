#!/bin/bash -eu
set -o pipefail

TA_NAME="Splunk_TA_otel"

if [ "$ARCH" == "amd64" ]; then
    SPLUNK_ARCH="x86_64"
else
    echo "Attempted to package unknown or unsupported architecture: $ARCH"
    exit 1
fi

TA_PACKAGING_DIR="$BUILD_DIR/out/$ARCH"
mkdir -p "$TA_PACKAGING_DIR/$TA_NAME"
# Copy config files into addon package directory
cp -rp "$BUILD_DIR/$TA_NAME/default" "$TA_PACKAGING_DIR/$TA_NAME/"
cp -rp "$BUILD_DIR/$TA_NAME/configs" "$TA_PACKAGING_DIR/$TA_NAME/"
cp -rp "$BUILD_DIR/$TA_NAME/README" "$TA_PACKAGING_DIR/$TA_NAME/"
cp -rp "$BUILD_DIR/$TA_NAME/README.md" "$TA_PACKAGING_DIR/$TA_NAME/"
cp -rp "$BUILD_DIR/$TA_NAME/static" "$TA_PACKAGING_DIR/$TA_NAME/"

# Copy collector into addon package directory
if [ "$PLATFORM" == "windows" ] || [ "$PLATFORM" == "all" ] ; then 
    COLLECTOR_BINARY="otelcol_windows_$ARCH.exe"
    cp -rp "$BUILD_DIR/$TA_NAME/windows_$SPLUNK_ARCH" "$TA_PACKAGING_DIR/$TA_NAME/"
    cp -rp "$BUILD_DIR/out/bin/$COLLECTOR_BINARY" "$TA_PACKAGING_DIR/$TA_NAME/windows_$SPLUNK_ARCH/bin/$COLLECTOR_BINARY"
fi
if [ "$PLATFORM" == "linux" ] || [ "$PLATFORM" == "all" ] ; then 
    COLLECTOR_BINARY="otelcol_linux_$ARCH"
    cp -rp "$BUILD_DIR/$TA_NAME/linux_$SPLUNK_ARCH" "$TA_PACKAGING_DIR/$TA_NAME/"
    cp -rp "$BUILD_DIR/out/bin/$COLLECTOR_BINARY" "$TA_PACKAGING_DIR/$TA_NAME/linux_$SPLUNK_ARCH/bin/$COLLECTOR_BINARY"
fi
if [ "$PLATFORM" == "darwin" ] ; then  # NOTE Darwin not used yet
    COLLECTOR_BINARY="otelcol_darwin_$ARCH"
    cp -rp "$BUILD_DIR/$TA_NAME/darwin_$SPLUNK_ARCH" "$TA_PACKAGING_DIR/$TA_NAME/"
    cp -rp "$BUILD_DIR/out/bin/$COLLECTOR_BINARY" "$TA_PACKAGING_DIR/$TA_NAME/darwin_$SPLUNK_ARCH/bin/$COLLECTOR_BINARY"
fi

# Copy smart agent bundle into addon package directory
version=""
if [ "$OTEL_COLLECTOR_VERSION" != "" ]; then 
    version="${OTEL_COLLECTOR_VERSION}_"
fi
if [ "$PLATFORM" == "windows" ] || [ "$PLATFORM" == "all" ] ; then
    cp "$BUILD_DIR/out/smart-agent/agent-bundle_${version}windows_${ARCH}.zip" "$TA_PACKAGING_DIR/$TA_NAME/windows_$SPLUNK_ARCH/bin/agent-bundle_windows_${ARCH}.zip"
fi
if [ "$PLATFORM" == "linux" ] || [ "$PLATFORM" == "all" ] ; then
    cp "$BUILD_DIR/out/smart-agent/agent-bundle_${version}linux_${ARCH}.tar.gz" "$TA_PACKAGING_DIR/$TA_NAME/linux_$SPLUNK_ARCH/bin/agent-bundle_linux_${ARCH}.tar.gz"
fi

# Copy some default discovery configuration examples
cp -R "$BUILD_DIR/configs/discovery" "$TA_PACKAGING_DIR/$TA_NAME/configs"
if [ "$PLATFORM" != "linux" ] && [ "$PLATFORM" != "all" ] ; then
  rm -rf "$TA_PACKAGING_DIR/$TA_NAME/configs/discovery/config.d.linux"
fi

# Prepare artifact directory structure
DEST_DIR="$BUILD_DIR/out/distribution/"
mkdir -p "$DEST_DIR"

# Package addon into artifact directory
OUT_DIR="$(realpath "$DEST_DIR")"
echo "creating tarball at $OUT_DIR/Splunk_TA_otel.tgz"
COPYFILE_DISABLE=1 tar -C "$TA_PACKAGING_DIR" -hcz --exclude '*.DS_Store' -f "$OUT_DIR/$TA_NAME.tgz" "$TA_NAME"
