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

# Set REMOVE_LEGACY_CONFIG=true when the package is explicitly being removed
# to delete the configuration backup created during migration. The command
# line option is useful when invoking this hook directly.
REMOVE_LEGACY_CONFIG="${REMOVE_LEGACY_CONFIG:-false}"
for option in "$@"; do
    if [ "$option" = "--remove-legacy-config" ]; then
        REMOVE_LEGACY_CONFIG=true
    fi
done

if [ "$REMOVE_LEGACY_CONFIG" = "true" ] && [ -d "$LEGACY_CONFIG_DIR" ]; then
    echo "Removing legacy configuration backup from $LEGACY_CONFIG_DIR"
    rm -rf "$LEGACY_CONFIG_DIR"
fi

if [ -f "$PRELOAD_PATH" ] && grep -q "$LIBOTELINJECT_PATH" "$PRELOAD_PATH"; then
    echo "Removing $LIBOTELINJECT_PATH from $PRELOAD_PATH"
    sed -i -e "s|$LIBOTELINJECT_PATH||" "$PRELOAD_PATH"
    if [ ! -s "$PRELOAD_PATH" ] || ! grep -q '[^[:space:]]' "$PRELOAD_PATH"; then
        echo "Removing empty $PRELOAD_PATH"
        rm -f "$PRELOAD_PATH"
    fi
fi
