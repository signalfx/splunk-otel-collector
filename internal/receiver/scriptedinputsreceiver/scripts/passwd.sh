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

PRINTF='END {printf "%s %s\n", DATE, FILEHASH}'
# shellcheck disable=SC2034
PASSWD_FILE=/etc/passwd

if [ "$KERNEL" = "Linux" ] || [ "$KERNEL" = "SunOS" ] || [ "$KERNEL" = "AIX" ] || [ "x$KERNEL" != "xHP-UX" ] || [ "$KERNEL" = "Darwin" ] || [ "$KERNEL" = "FreeBSD" ] ; then
	assertHaveCommand date
	# shellcheck disable=SC2016
    CMD='eval date ; eval LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256 $PASSWD_FILE ; cat $PASSWD_FILE'
	# shellcheck disable=SC2016
	PARSE_0='NR==1 {DATE=$0}'
	# shellcheck disable=SC2016
	PARSE_1='NR==2 {FILEHASH="file_hash=" $2}'
	# Note the inline print in the next PARSE statement.
	# Comments are eliminated from the output, but included in FILEHASH.
	# shellcheck disable=SC2016
	PARSE_2='NR>2 && /^[^#]/ { split($0, arr, ":") ; printf "%s user=%s password=x user_id=%s user_group_id=%s home=%s shell=%s\n", DATE, arr[1], arr[3], arr[4], arr[6], arr[7]}'

	MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2"

fi

$CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $PRINTF"
echo "Cmd = [$CMD];  | $AWK '$MASSAGE $PRINTF'" >> "$TEE_DEST"
