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

PRELOAD_PATH="/etc/ld.so.preload"
LIBOTELINJECT_PATH="/usr/lib/splunk-instrumentation/libotelinject.so"
LEGACY_CONFIG_DIR="/usr/lib/splunk-instrumentation/legacy-zeroconfig"

# This script runs as the Debian prerm/RPM %preun hook (fpm --before-remove), which fires on a
# package upgrade as well as an actual uninstall. dpkg passes "upgrade"/"failed-upgrade" (vs.
# "remove") as $1 to prerm; rpm passes the count of package instances that will remain after this
# action to %preun ("0" on final removal, ">=1" mid-upgrade). Only strip the preload entry on a
# real removal: during an upgrade, libotelinject.so's path is unchanged between versions, so
# leaving the entry in place is correct, and nothing else re-adds it afterwards.
ACTION="${1:-remove}"
case "$ACTION" in
    upgrade|failed-upgrade|deconfigure)
        IS_UNINSTALL=false
        ;;
    [0-9]*)
        if [ "$ACTION" -ge 1 ] 2>/dev/null; then
            IS_UNINSTALL=false
        else
            IS_UNINSTALL=true
        fi
        ;;
    *)
        IS_UNINSTALL=true
        ;;
esac

# Set REMOVE_LEGACY_CONFIG=true when the package is explicitly being removed
# to delete the configuration backup created during migration. The command
# line option is useful when invoking this hook directly. Gated on IS_UNINSTALL
# so that leaving this flag set doesn't wipe the backup on a routine upgrade.
REMOVE_LEGACY_CONFIG="${REMOVE_LEGACY_CONFIG:-false}"
for option in "$@"; do
    if [ "$option" = "--remove-legacy-config" ]; then
        REMOVE_LEGACY_CONFIG=true
    fi
done

if [ "$IS_UNINSTALL" = "true" ] && [ "$REMOVE_LEGACY_CONFIG" = "true" ] && [ -d "$LEGACY_CONFIG_DIR" ]; then
    echo "Removing legacy configuration backup from $LEGACY_CONFIG_DIR"
    rm -rf "$LEGACY_CONFIG_DIR"
fi

if [ "$IS_UNINSTALL" = "true" ] && [ -f "$PRELOAD_PATH" ] && grep -q "$LIBOTELINJECT_PATH" "$PRELOAD_PATH"; then
    echo "Removing $LIBOTELINJECT_PATH from $PRELOAD_PATH"
    sed -i -e "s|$LIBOTELINJECT_PATH||" "$PRELOAD_PATH"
    if [ ! -s "$PRELOAD_PATH" ] || ! grep -q '[^[:space:]]' "$PRELOAD_PATH"; then
        echo "Removing empty $PRELOAD_PATH"
        rm -f "$PRELOAD_PATH"
    fi
fi
