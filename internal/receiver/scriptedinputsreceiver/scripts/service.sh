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

if [ "$KERNEL" = "Linux" ] ; then
	if ! queryHaveCommand systemctl; then
		assertHaveCommand date
		assertHaveCommand chkconfig
		CMD='eval date ; /sbin/chkconfig --list'
		# shellcheck disable=SC2016
		PARSE_0='NR==1 {DATE=$0 ; FULLTEXT=""}'
		# shellcheck disable=SC2016
		PARSE_1='NR>1 {
			FULLTEXT = FULLTEXT $0 "\n"
			split($0, ARR)
			EVT="app=" ARR[1]
			for (i=0 ; i<7 ; i++) {
				split(ARR[i+2], STATE, ":")
				EVT = EVT " runlevel" i "=" STATE[2]
			}
			if (ARR[1] ~ /[Ss][Pp][Ll][Uu][Nn][Kk]/) { SPLUNKD=1 }
			printf "%s type=chkconfig %s\n", DATE, EVT
		}'
		MASSAGE="$PARSE_0 $PARSE_1"

		# Send the collected full text to openssl; this avoids any timing discrepancies
		# between when the information is collected and when we process it.
		# shellcheck disable=SC2016
		POSTPROCESS='END {
			if (SPLUNKD==0) { printf "%s app=\"Splunk\" StartMode=Disabled\n", DATE }
			printf "%s %s", DATE, "file_hash="
			printf "%s", FULLTEXT | "LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256"
		}'
	else
		assertHaveCommand systemctl
		assertHaveCommand date

		# Run the systemctl command to get all units and their state
		CMD='eval date; systemctl list-units --type=service --all'
		# shellcheck disable=SC2016
		PARSE_0='NR==1 {DATE=$0}'
		# shellcheck disable=SC2016
		PARSE_1='
			# On header row, get lengths to the fields
			 NR==2  {
				match($0, /^ */); leading=RLENGTH;
				match($0, /^.*DESC/); desclen=RLENGTH-4;
				FULLTEXT="";
				next;
			}'
		# shellcheck disable=SC2016
		PARSE_2='(NR > 2){
			# Stop at the empty line
			if ( !NF ) { exit; }
			# Skip the leading spaces
			$0 = substr( $0, leading );
			# the description spans fields so catch it seperately
			desc=substr( $0, desclen );
			FULLTEXT = FULLTEXT $0 "\n"
			if ($1 ~ /[Ss][Pp][Ll][Uu][Nn][Kk]/) { SPLUNKD=1 }
			printf "%s type=systemctl UNIT=%s, LOADED=%s, ACTIVE=%s, SUB=%s, DESCRIPTION=\"%s\" \n",DATE, $1, $2, $3, $4, desc
		}'
		MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2"
		# shellcheck disable=SC2016
		POSTPROCESS='END {
			if (SPLUNKD==0) { printf "%s app=\"Splunk\" StartMode=Disabled\n", DATE }
			printf "%s %s", DATE, "file_hash="
			printf "%s", FULLTEXT | "LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256"
		}'
	fi

elif [ "$KERNEL" = "SunOS" ] ; then
	assertHaveCommand date
	assertHaveCommand svcs

	CMD='eval date ; svcs -H -a -o STATE,FMRI'
	# shellcheck disable=SC2016
	PARSE_0='NR==1 {DATE=$0 ; FULLTEXT=""}'
	# shellcheck disable=SC2016
	PARSE_1='NR>1 {
		STATE="State=\""$1"\""
		idx=index($2,":")
		STARTNAME="StartName=\""substr($2,0,idx-1)"\""
		APP="app=\""substr($2,idx+1)"\""
		FULLTEXT=FULLTEXT $0 "\n"
	}'
	PARSE_2='/^legacy_run/ {
		STARTMODE="StartMode=\"Auto\""
	}'
	PARSE_3='/^online/ {
		STARTMODE="StartMode=\"Auto\""
		STATE="State=\"Running\""
	}'
	PARSE_4='/^disabled/ {
		STARTMODE="StartMode=\"Disabled\""
		STATE="State=\"Stopped\""
	}'

	INLINE_PRINT='NR>1 && APP!=0 {printf "%s %s %s %s %s\n", DATE, APP, STARTMODE, STARTNAME, STATE}'

	MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3 $PARSE_4 $INLINE_PRINT"

	# Send the collected full text to openssl; this avoids any timing discrepancies
	# between when the information is collected and when we process it.
	# shellcheck disable=SC2016
	POSTPROCESS='END {
		if (SPLUNKD==0) { printf "%s app=\"Splunk\" StartMode=Disabled\n", DATE }
		printf "%s %s", DATE, "file_hash="
		printf "%s", FULLTEXT | "LD_LIBRARY_PATH=$SPLUNK_HOME/lib $SPLUNK_HOME/bin/openssl sha256"
	}'

elif [ "$KERNEL" = "Darwin" ] ; then

	assertHaveCommand date
	assertHaveCommand defaults
	assertHaveCommand dscl
	assertHaveCommand find
	assertHaveCommand ls

	# Get startup items
	CMD='eval date ; ls -1 /System/Library/StartupItems/ /Library/StartupItems/'
	# Get per-user startup items
	# shellcheck disable=SC2044
	for PLIST_FILE in $(find /Users -name "loginwindow.plist") ; do
		CMD=$CMD' ; echo '$PLIST_FILE': ; defaults read '$PLIST_FILE
	done
	# shellcheck disable=SC2016
	PARSE_0='NR==1 {DATE=$0}'
	# Retrieve path for system startup items
	# shellcheck disable=SC2016
	PARSE_1='/^\/(System|Library)/ {
		split($0, tmparr, ":")
		PATH="file_path=\""tmparr[1]
		USER=0
		START_MODE="StartMode=Auto"
		START_TYPE="StartType=startup"
	}'

	# Retrieve user information for user startup items.
	# shellcheck disable=SC2016
	PARSE_2='/^\/Users/ {
		split($0, tmparr, "/")
		USER="user=" tmparr[3]
		START_MODE="StartMode=Auto"
		START_TYPE="StartType=login"
	}'

	# Retrieve the path for user startup items.
	# shellcheck disable=SC2016
	PARSE_3='/[[:blank:]]*Path/ {
		split($0, path_arr, "=")
		num=split(path_arr[2], app_arr, "/")
		split(app_arr[num], app_final, ".")
		split(path_arr[2], path_final, "\"")
		APP="app=\"" app_final[1] "\""
		FILE_PATH="file_path=\"" path_final[2] "\""

		# Only print if we find a path.
		printf "%s %s %s %s %s %s\n", DATE, APP, START_MODE, START_TYPE, FILE_PATH, USER

		# Note that we found splunkd if app matches
		if (APP ~ /[Ss][Pp][Ll][Uu][Nn][Kk]/) { SPLUNKD=1 }
	}'

	# Retrieve the system startup item name from the output of "ls -1"
	# shellcheck disable=SC2016
	PARSE_4='/^[^\/]/ {
		if (NR>1 && USER==0 && NF > 0) {
			APP="app=\""$0"\""
			PATH=PATH$0"\""
			printf "%s %s %s %s %s\n", DATE, APP, START_MODE, START_TYPE, PATH
		}

		# Note that we found splunkd if app matches
		if (APP ~ /[Ss][Pp][Ll][Uu][Nn][Kk]/) { SPLUNKD=1 }

	}'

	MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3 $PARSE_4"

	POSTPROCESS='END { if (SPLUNKD==0) { printf "%s app=\"Splunk\" StartMode=Disabled\n", DATE } }'

else
	# Exits
	failUnsupportedScript
fi

$CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $POSTPROCESS"
echo "Cmd = [$CMD];  | $AWK '$MASSAGE $POSTPROCESS'" >> "$TEE_DEST"
