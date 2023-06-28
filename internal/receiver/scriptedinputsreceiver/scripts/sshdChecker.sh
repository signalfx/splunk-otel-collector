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

SSH_CONFIG_FILE=""
if [ "$KERNEL" = "Linux" ] || [ "$KERNEL" = "SunOS" ] ; then
	SSH_CONFIG_FILE=/etc/ssh/sshd_config
elif [ "$KERNEL" = "Darwin" ] ; then
	SSH_CONFIG_FILE=/etc/sshd_config
else
	failUnsupportedScript
fi

FILL_BLANKS='END {
	if (SSHD_PROTOCOL == 0) {
		SSHD_PROTOCOL=SSHD_DEFAULT_PROTOCOL
	}'

PRINTF='{printf "%s app=sshd %s %s\n", DATE, FILEHASH, SSHD_PROTOCOL}}'

if [ "x$SOLARIS_11" != "xtrue" ] ; then

	# If $SSH_CONFIG_FILE file exists and is a regular file.
	if [ -f "$SSH_CONFIG_FILE" ] ; then

		assertHaveCommand cat

		# Get file hash
		# shellcheck disable=SC2016
		CMD='eval date ; eval LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256 $SSH_CONFIG_FILE ; cat $SSH_CONFIG_FILE'

		# Get the date.
		# shellcheck disable=SC2016
		PARSE_0='NR==1 {DATE=$0}'

		# Try to use cross-platform case-insensitive matching for text. Note
		# that "match", "tolower", IGNORECASE and other common awk commands or
		# options are actually nawk/gawk extensions so avoid them if possible.
		# shellcheck disable=SC2016
		PARSE_1='/^[Pp][Rr][Oo][Tt][Oo][Cc][Oo][Ll]/ {
			split($0, arr)
			num = split(arr[2], protocols, ",")
			if (num == 2) {
				SSHD_PROTOCOL="sshd_protocol=" protocols[1] "/" protocols[2]
			} else {
				SSHD_PROTOCOL="sshd_protocol=" protocols[1]
			}
		}'
		# shellcheck disable=SC2016
		PARSE_2='/^#[[:blank:]]*[Pp][Rr][Oo][Tt][Oo][Cc][Oo][Ll]/ {
			num=split($0, arr)
			protonum = split(arr[num], protocols, ",")
			if (protonum == 2) {
				SSHD_DEFAULT_PROTOCOL="sshd_protocol=" protocols[1] "/" protocols[2]
			} else {
				SSHD_DEFAULT_PROTOCOL="sshd_protocol=" protocols[1]
			}
		}'
		# shellcheck disable=SC2016
		PARSE_3='/^SHA256/ {FILEHASH="file_hash=" $2}'

		MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3"

	else
		# shellcheck disable=SC2016
		echo "SSHD configuration (file: $SSH_CONFIG_FILE) missing or unreadable." >> "$TEE_DEST"
		exit 1
	fi

else

	if [ -f "$SSH_CONFIG_FILE" ] && [ -r "$SSH_CONFIG_FILE" ] ; then

		# Solaris 11 only supports SSH protocol 2.
		assertHaveCommand cat

		# Get file hash
		# shellcheck disable=SC2016
		CMD='eval date ; eval LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256 $SSH_CONFIG_FILE'
		# shellcheck disable=SC2016
		PARSE_0='NR==1 {DATE=$0 ; SSHD_PROTOCOL="sshd_protocol=2"}'
		# shellcheck disable=SC2016
		PARSE_1='/^SHA256/ {FILEHASH="file_hash=" $2}'

		MASSAGE="$PARSE_0 $PARSE_1"

	else
		echo "SSHD configuration (file: $SSH_CONFIG_FILE) missing or unreadable." >> "$TEE_DEST"
		exit 1
	fi

fi

$CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $FILL_BLANKS $PRINTF"
echo "Cmd = [$CMD];  | $AWK '$MASSAGE $FILL_BLANKS $PRINTF'" >> "$TEE_DEST"
