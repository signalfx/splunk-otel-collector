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
LIBSPLUNK_PATH="/usr/lib/splunk-instrumentation/libsplunk.so"

if [ -f "$PRELOAD_PATH" ]; then
    if ! grep -q "$LIBSPLUNK_PATH" "$PRELOAD_PATH"; then
        # backup existing file with timestamp
        if command -v rpm >/dev/null 2>&1; then
            # need to escape '%' for rpm macros
            ts="$(date '+%%Y%%m%%d-%%H%%M%%S')"
        else
            ts="$(date '+%Y%m%d-%H%M%S')"
        fi
        echo "Saving $PRELOAD_PATH as ${PRELOAD_PATH}.bak.${ts}"
        cp "$PRELOAD_PATH" "${PRELOAD_PATH}.bak.${ts}"
        echo "Adding $LIBSPLUNK_PATH to $PRELOAD_PATH"
        echo "$LIBSPLUNK_PATH" >> "$PRELOAD_PATH"
    fi
else
    echo "Adding $LIBSPLUNK_PATH to $PRELOAD_PATH"
    echo "$LIBSPLUNK_PATH" >> "$PRELOAD_PATH"
fi
