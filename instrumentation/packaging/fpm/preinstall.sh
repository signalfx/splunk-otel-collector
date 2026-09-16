#!/bin/sh

# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PKG_NAME="splunk-otel-auto-instrumentation"
LIBSPLUNK_PATH="/usr/lib/splunk-instrumentation/libsplunk.so"
LEGACY_CONFIG_DIR="/usr/lib/splunk-instrumentation/legacy-zeroconfig"
ZEROCONFIG_DIR="/etc/splunk/zeroconfig"
PRELOAD_PATH="/etc/ld.so.preload"

# Detect an upgrade from a legacy (pre-injector) install by checking for the
# old libsplunk.so file on disk rather than querying the installed package
# version: on rpm, "rpm -q" inside %pre already resolves to the *incoming*
# package during an upgrade transaction, not the previously installed one,
# so version-based detection is unreliable there. The file-based check works
# the same way for both deb (preinst) and rpm (%pre).
if [ -f "$LIBSPLUNK_PATH" ]; then
    echo "WARNING: Upgrading $PKG_NAME from a version using libsplunk.so. Auto-instrumentation is switching from libsplunk.so to libotelinject.so, and configuration files have moved from /etc/splunk/zeroconfig/ to /etc/opentelemetry/injector/. See the release notes for details." >&2

    # Preserve the legacy configuration files alongside libotelinject.so
    # before removing the old configuration directory. The files may contain
    # operator-provided settings that are useful when completing the migration.
    if [ -d "$ZEROCONFIG_DIR" ]; then
        mkdir -p "$LEGACY_CONFIG_DIR"
        cp -a "$ZEROCONFIG_DIR"/. "$LEGACY_CONFIG_DIR"/
    fi

    # The legacy package's config files are declared as conffiles/%config,
    # so neither dpkg nor rpm remove them automatically on upgrade when the
    # new package no longer ships them. Remove them here so they don't
    # linger on disk after the switch to libotelinject.so.
    rm -rf "$ZEROCONFIG_DIR"

    # libsplunk.so itself is removed automatically once the transaction
    # completes, since it is not shipped by the new package and is not a
    # conffile. But nothing else in this package rewrites /etc/ld.so.preload
    # on a plain package-manager upgrade (that's left to config-management
    # tools/the installer script, which fully re-render the file), so strip
    # the stale entry here to avoid leaving a dangling reference to a
    # deleted library in the interim.
    if [ -f "$PRELOAD_PATH" ] && grep -q "$LIBSPLUNK_PATH" "$PRELOAD_PATH"; then
        echo "Removing $LIBSPLUNK_PATH from $PRELOAD_PATH"
        sed -i -e "s|$LIBSPLUNK_PATH||" "$PRELOAD_PATH"
        if [ ! -s "$PRELOAD_PATH" ] || ! grep -q '[^[:space:]]' "$PRELOAD_PATH"; then
            echo "Removing empty $PRELOAD_PATH"
            rm -f "$PRELOAD_PATH"
        fi
    fi
fi
