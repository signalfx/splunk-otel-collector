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

# In AWK scripts in this file, the following are true:
# 	FULLTEXT is used to capture the output for SHA256 checksum generation.
# 	SPLUNKD is used to determine Splunk service status.

if [ "$KERNEL" = "Linux" ] || [ "$KERNEL" = "Darwin" ] ; then
	assertHaveCommand date
	assertHaveCommand lsof
	if [ -f /usr/sbin/lsof ] ; then
		LSOF=/usr/sbin/lsof
	elif [ -f /usr/bin/lsof ] ; then
		# shellcheck disable=SC2034
		LSOF=/usr/bin/lsof
	fi
	# shellcheck disable=SC2016
	CMD='eval date ; ${LSOF} -i -P -n'
	# shellcheck disable=SC2016
	PARSE_0='NR==1 {DATE=$0 ; FULLTEXT=""}'
	# Only base the file hash on the listening ports, not on
	# open connections.
	# shellcheck disable=SC2016
	PARSE_1='/LISTEN|[Uu][Dd][Pp]/ {
		FULLTEXT = FULLTEXT $0 "\n"
		idx=match($0, /\(LISTEN\)/)
		if (idx>0) {
			DATA=substr($0, 0, idx-1)
		} else {
			DATA=$0
		}
		fields = split(DATA, portarr)

		# This compensates for varying field counts.
		if (fields == 9) {
			hostfields = split(portarr[9], hostarr, ":")
			TRANSPORT="transport=" portarr[8]
		} else if (fields == 8) {
			hostfields = split(portarr[8], hostarr, ":")
			TRANSPORT="transport=" portarr[7]
		}

		if (hostfields == 2 && hostarr[2] ~ /[0-9][0-9]*/) {
			DESTIP="dest_ip=" hostarr[1]
			DESTPORT="dest_port=" hostarr[2]
			APP="app=" portarr[1]
			PID="pid=" portarr[2]
			USER="user=" portarr[3]
			FD="fd=" portarr[4]
			IPVERSION="ip_version=" substr(portarr[5],index(portarr[5],"v")+1)
			DVCID="dvc_id=" portarr[6]
			#printf "MATCH: %s\n", $0
			printf "%s %s %s %s %s %s %s %s %s %s\n", DATE, APP, DESTIP, DESTPORT, PID, USER, FD, IPVERSION, DVCID, TRANSPORT
		} else {
			#printf "NOMATCH: %s\n", $0
			;
		}
	}'
	MASSAGE="$PARSE_0 $PARSE_1"

	# Send the collected full text to openssl; this avoids any timing discrepancies
	# between when the information is collected and when we process it.
	# shellcheck disable=SC2016
	POSTPROCESS='END {
		printf "%s %s", DATE, "file_hash="
		printf "%s", FULLTEXT | "LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256"
	}'

elif [ "$KERNEL" = "SunOS" ] ; then

	assertHaveCommand date
	assertHaveCommand netstat

	CMD='eval date ; netstat -an -f inet -f inet6'
	# shellcheck disable=SC2016
	PARSE_0='NR==1 {DATE=$0 ; FULLTEXT=""}'
	# shellcheck disable=SC2016
	PARSE_1='/^[Tt][Cc][Pp]|[Uu][Dd][Pp]/ {
                split($0, protoarr, ":")
                TRANSPORT="transport=" protoarr[1]
                IPVERSION="ip_version=" substr(protoarr[2],index(protoarr[2],"v")+1)
                next
        }'
	# shellcheck disable=SC2016
	PARSE_3='NR>1 && $0 !~ /Local|^-|^$/ {
		FULLTEXT = FULLTEXT $0 "\n"
		split($0, arr)
		num = split(arr[1], hostarr, "\.")
		if ( TRANSPORT ~ /[Tt][Cc][Pp]/) {
			DESTIP="dest_ip="hostarr[1]
		} else {
			DESTIP="dest_dns="hostarr[1]
		}
		DESTPORT=hostarr[num]

		for (i=2; i<num; i++) {
			DESTIP=DESTIP"."hostarr[i]
		}
		if ( $0 !~ /[Uu][Nn][Bb][Oo][Uu][Nn][Dd]/ && DESTPORT != "*" ) {
			DESTPORT="dest_port="DESTPORT
			printf "%s %s %s %s %s \n", DATE, DESTIP, DESTPORT, IPVERSION, TRANSPORT
		}
	}'

	MASSAGE="$PARSE_0 $PARSE_1 $PARSE_3"

	# Send the collected full text to openssl; this avoids any timing discrepancies
	# between when the information is collected and when we process it.
	# shellcheck disable=SC2016
	POSTPROCESS='END {
		printf "%s %s", DATE, "file_hash="
		printf "%s", FULLTEXT | "LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256"
	}'

else
	# Exits
	failUnsupportedScript
fi

$CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $POSTPROCESS"
echo "Cmd = [$CMD];  | $AWK '$MASSAGE $POSTPROCESS'" >> "$TEE_DEST"
