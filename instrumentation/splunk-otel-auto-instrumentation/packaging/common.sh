#!/bin/bash

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

CUR_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"
FPM_DIR="$( cd "$CUR_DIR/../../packaging/fpm" && pwd )"
. $FPM_DIR/common.sh

PKG_NAME="splunk-otel-auto-instrumentation"

LIBSPLUNK_INSTALL_PATH="/usr/lib/splunk-instrumentation/libsplunk.so"
CONFIG_INSTALL_PATH="/usr/lib/splunk-instrumentation/instrumentation.conf"

POSTINSTALL_PATH="$CUR_DIR/postinstall.sh"
PREUNINSTALL_PATH="$CUR_DIR/preuninstall.sh"
CONFIG_PATH="$CUR_DIR/../install/instrumentation.conf"

setup_files_and_permissions() {
    local libsplunk="$1"
    local java_agent="$2"
    local buildroot="$3"

    mkdir -p "$buildroot/$(dirname $LIBSPLUNK_INSTALL_PATH)"
    cp -f "$libsplunk" "$buildroot/$LIBSPLUNK_INSTALL_PATH"
    sudo chown root:root "$buildroot/$LIBSPLUNK_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$LIBSPLUNK_INSTALL_PATH"

    mkdir -p "$buildroot/$(dirname $JAVA_AGENT_INSTALL_PATH)"
    cp -f "$java_agent" "$buildroot/$JAVA_AGENT_INSTALL_PATH"
    sudo chown root:root "$buildroot/$JAVA_AGENT_INSTALL_PATH"
    sudo chmod 755 "$buildroot/$JAVA_AGENT_INSTALL_PATH"

    mkdir -p "$buildroot/$(dirname $CONFIG_INSTALL_PATH)"
    cp -f "$CONFIG_PATH" "$buildroot/$CONFIG_INSTALL_PATH"
    sudo chown root:root "$buildroot/$CONFIG_INSTALL_PATH"
    sudo chmod 644 "$buildroot/$CONFIG_INSTALL_PATH"
}
