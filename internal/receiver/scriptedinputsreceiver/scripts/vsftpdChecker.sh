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

# VSFTPD configuration file format is common to all platforms, but may be in one
# of several locations (and may also be restricted to root).
if [ -f /etc/vsftpd.conf ] ; then
	VSFTPD_CONFIG_FILE=/etc/vsftpd.conf
elif [ -f /etc/vsftpd/vsftpd.conf ] ; then
	VSFTPD_CONFIG_FILE=/etc/vsftpd/vsftpd.conf
elif [ -f /private/etc/vsftpd.conf ] ; then
	# Usually MAC OS X
	VSFTPD_CONFIG_FILE=/private/etc/vsftpd.conf
elif [ -f /usr/local/etc/vsftpd.conf ] ; then
        # To support MAC OS 10.15
        VSFTPD_CONFIG_FILE=/usr/local/etc/vsftpd.conf
fi

# Set the default. If the file is readable and has "anonymous_enable" commented
# out, the default behavior is to ALLOW anonymous FTP. Reset the value of
# anonymous_enable in the output if this is the case
# line, then the allowed protocols will be the default of "2,1".
FILL_BLANKS='END {
	if (ANON_DEFAULT != 0) {
		ANON_ENABLE=ANON_DEFAULT
	}'
PRINTF='{printf "%s app=vsftp %s %s %s\n", DATE, FILEHASH, LOCAL_ENABLE, ANON_ENABLE}}'

# If $VSFTPD_CONFIG_FILE file exists and is a regular file.
if [ -f "$VSFTPD_CONFIG_FILE" ] ; then

        assertHaveCommand cat
        assertHaveCommand date

        # Get file hash
        # shellcheck disable=SC2016
        CMD='eval date ; eval LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256 $VSFTPD_CONFIG_FILE ; cat $VSFTPD_CONFIG_FILE'

        # Get the date.
        # shellcheck disable=SC2016
        PARSE_0='NR==1 {DATE=$0}'

        # Try to use cross-platform case-insensitive matching for text. Note
        # that "match", "tolower", IGNORECASE and other common awk commands or
        # options are actually nawk/gawk extensions so avoid them if possible.
        # shellcheck disable=SC2016
        PARSE_1='/[Ll][Oo][Cc][Aa][Ll][_][Ee][Nn][Aa][Bb][Ll][Ee]/ { split($0, arr, "=") ; LOCAL_ENABLE="local_enable=" arr[2] } '
        # shellcheck disable=SC2016
        PARSE_2='/^[Aa][Nn][Oo][Nn][Yy][Mm][Oo][Uu][Ss][_][Ee][Nn][Aa][Bb][Ll][Ee]/ { split($0, arr, "=") ; ANON_ENABLE="anonymous_enable=" arr[2] } '
        # The default behavior is to permit anonymous FTP
        PARSE_3='/^[#]+[[:blank:]]*[Aa][Nn][Oo][Nn][Yy][Mm][Oo][Uu][Ss][_][Ee][Nn][Aa][Bb][Ll][Ee]/ { ANON_DEFAULT="anonymous_enable=YES"} '
        # shellcheck disable=SC2016
        PARSE_4='/^SHA256/ {FILEHASH="file_hash=" $2}'

        MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3 $PARSE_4"

        $CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $FILL_BLANKS $PRINTF"
        echo "Cmd = [$CMD];  | $AWK '$MASSAGE $FILL_BLANKS $PRINTF'" >> "$TEE_DEST"

else
		echo "VSFTPD configuration file not found." >> "$TEE_DEST"
fi
