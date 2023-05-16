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

if [ -f "$PRELOAD_PATH" ] && grep -q "$LIBSPLUNK_PATH" "$PRELOAD_PATH"; then
    echo "Removing $LIBSPLUNK_PATH from $PRELOAD_PATH"
    sed -i -e "s|$LIBSPLUNK_PATH||" "$PRELOAD_PATH"
    if [ ! -s "$PRELOAD_PATH" ] || ! grep -q '[^[:space:]]' "$PRELOAD_PATH"; then
        echo "Removing empty $PRELOAD_PATH"
        rm -f "$PRELOAD_PATH"
    fi
fi
