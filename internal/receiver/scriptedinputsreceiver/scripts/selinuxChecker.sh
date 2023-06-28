#!/bin/sh
# Copyright The OpenTelemetry Authors
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
#

# shellcheck disable=SC1091
. "$(dirname "$0")"/common.sh

PRINTF='END {printf "%s app=selinux %s %s %s %s\n", DATE, FILEHASH, SELINUX, SELINUXTYPE, SETLOCALDEFS}'

if [ "$KERNEL" = "Linux" ] ; then
        if [ -f /etc/sysconfig/selinux ] ; then
           SELINUX_FILE=/etc/sysconfig/selinux
        elif [ -f /etc/selinux/config ] ; then
        # shellcheck disable=SC2034
           SELINUX_FILE=/etc/selinux/config
        else
              echo "SELinux not configured."  >> "$TEE_DEST"
              exit 1
        fi

        assertHaveCommand cat

        # Get file hash
        # shellcheck disable=SC2016
        CMD='eval date ; eval LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256 $SELINUX_FILE ; cat $SELINUX_FILE'

        # Get the date.
        # shellcheck disable=SC2016
        PARSE_0='NR==1 {DATE=$0}'

        # Try to use cross-platform case-insensitive matching for text. Note
        # that "match", "tolower", IGNORECASE and other common awk commands or
        # options are actually nawk/gawk extensions so avoid them if possible.
        # shellcheck disable=SC2016
        PARSE_1='/^[Ss][Ee][Ll][Ii][Nn][Uu][Xx]\=/ { SELINUX="selinux=" substr($0,index($0,"=")+1,length($0)) } '
        # shellcheck disable=SC2016
        PARSE_2='/^[Ss][Ee][Ll][Ii][Nn][Uu][Xx][Tt][Yy][Pp][Ee]\=/ { SELINUXTYPE="selinuxtype=" substr($0,index($0,"=")+1,length($0)) } '
        # shellcheck disable=SC2016
        PARSE_3='/^[Ss][Ee][Tt][Ll][Oo][Cc][Aa][Ll][Dd][Ee][Ff][Ss]\=/ { SETLOCALDEFS="setlocaldefs=" substr($0,index($0,"=")+1,length($0)) } '
        # shellcheck disable=SC2016
        PARSE_4='/^SHA256/ {FILEHASH="file_hash=" $2}'

        MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3 $PARSE_4"

        $CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $PRINTF"
        echo "Cmd = [$CMD];  | $AWK '$MASSAGE $PRINTF'" >> "$TEE_DEST"

fi
