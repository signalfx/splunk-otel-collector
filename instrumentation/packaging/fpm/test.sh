#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
. $SCRIPT_DIR/common.sh

PKG_TYPE="${1:-rpm}"
ARCH="${2:-amd64}"
INSTALL_CMD="rpm -ivh"
UNINSTALL_CMD="rpm -evh"
UPGRADE_CMD="rpm -Uvh --replacepkgs"

if [[ "$PKG_TYPE" = "rpm" ]]; then
    if [[ "$ARCH" = "arm64" ]]; then
        ARCH="aarch64"
    elif [[ "$ARCH" = "amd64" ]]; then
        ARCH="x86_64"
    fi
else
    INSTALL_CMD="dpkg -i"
    UNINSTALL_CMD="dpkg -P"
    UPGRADE_CMD="dpkg -i"
fi

PKG_PATH="${REPO_DIR}/instrumentation/dist/${PKG_NAME}*${ARCH}.${PKG_TYPE}"

# install
echo "Installing $PKG_PATH"
$INSTALL_CMD $PKG_PATH
if [ ! -f /etc/ld.so.preload ]; then
    echo "/etc/ld.so.preload not found" >&2
    exit 1
fi
if ! grep -q "$LIBSPLUNK_INSTALL_PATH" /etc/ld.so.preload; then
    echo "$LIBSPLUNK_INSTALL_PATH not found in /etc/ld.so.preload" >&2
    exit 1
fi

ldd "$LIBSPLUNK_INSTALL_PATH"

# upgrade
echo "Upgrading $PKG_PATH"
$UPGRADE_CMD $PKG_PATH
if [ ! -f /etc/ld.so.preload ]; then
    echo "/etc/ld.so.preload not found" >&2
    exit 1
fi
if ! grep -q "$LIBSPLUNK_INSTALL_PATH" /etc/ld.so.preload; then
    echo "$LIBSPLUNK_INSTALL_PATH not found in /etc/ld.so.preload" >&2
    exit 1
fi

ldd "$LIBSPLUNK_INSTALL_PATH"

# uninstall
echo "Uninstalling $PKG_NAME"
$UNINSTALL_CMD "$PKG_NAME"
if [[ -f /etc.ld.preload ]] && grep -q "$LIBSPLUNK_INSTALL_PATH" /etc/ld.so.preload; then
    echo "$LIBSPLUNK_INSTALL_PATH not removed from /etc/ld.so.preload" >&2
    exit 1
fi

# install with pre-existing /etc/ld.so.preload
touch /etc/ld.so.preload
echo "Installing $PKG_PATH"
$INSTALL_CMD $PKG_PATH
if [ ! -f /etc/ld.so.preload ]; then
    echo "/etc/ld.so.preload not found" >&2
    exit 1
fi
if ! grep -q "$LIBSPLUNK_INSTALL_PATH" /etc/ld.so.preload; then
    echo "$LIBSPLUNK_INSTALL_PATH not found in /etc/ld.so.preload" >&2
    exit 1
fi
for backup in /etc/ld.so.preload.bak.*; do
    if [[ -f "$backup" ]]; then
        echo "Found $backup"
    else
        echo "Backup not created for /etc/ld.so.preload" >&2
        exit 1
    fi
done

ldd "$LIBSPLUNK_INSTALL_PATH"

# uninstall
echo "Uninstalling $PKG_NAME"
$UNINSTALL_CMD "$PKG_NAME"
if [[ -f /etc.ld.preload ]] && grep -q "$LIBSPLUNK_INSTALL_PATH" /etc/ld.so.preload; then
    echo "$LIBSPLUNK_INSTALL_PATH not removed from /etc/ld.so.preload" >&2
    exit 1
fi
