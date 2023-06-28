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

if [ "$KERNEL" = "Linux" ] ; then
	assertHaveCommand date
    OSName=$(cat /etc/*release | grep '\bNAME=' | cut -d '=' -f2 | tr ' ' '_' | cut -d\" -f2)
	# Ubuntu doesn't have yum installed by default hence apt is being used to get the list of upgradable packages
    if [ "$OSName" = "Ubuntu" ]; then
		assertHaveCommand apt
		assertHaveCommand sed
		# sed command here replaces '/, [, ]' with ' '
		CMD='eval date ; eval apt list --upgradable | sed "s/\// /; s/\[/ /; s/\]/ /"'
		# shellcheck disable=SC2016
		PARSE_0='NR==1 {DATE=$0}'
		# shellcheck disable=SC2016
		PARSE_1='NR>2 { printf "%s package=%s ubuntu_update_stream=%s latest_package_version=%s ubuntu_architecture=%s current_package_version=%s\n", DATE, $1, $2, $3, $4, $7}'
		MESSAGE="$PARSE_0 $PARSE_1"
	else
		assertHaveCommand yum

		CMD='eval date ; yum check-update'
		# shellcheck disable=SC2016
		PARSE_0='NR==1 {
			DATE=$0
			PROCESS=0
			UPDATES["addons"]=0
			UPDATES["base"]=0
			UPDATES["extras"]=0
			UPDATES["updates"]=0
		}'

		# Skip extraneous text up to first blank line.
		# shellcheck disable=SC2016
		PARSE_1='NR>1 && PROCESS==0 && $0 ~ /^[[:blank:]]*$|^$/ {
			PROCESS=1
		}'
		# shellcheck disable=SC2016
		PARSE_2='NR>1 && PROCESS==1 {
			num = split($0, update_array)
			if (num == 3) {
				# Record the update count
				UPDATES[update_array[3]] = UPDATES[update_array[3]]+1
				printf "%s package=\"%s\" package_type=\"%s\"\n", DATE, update_array[1], update_array[3]
			} else if (num==2 && update_array[1] != "") {
				printf "%s package=\"%s\"\n", DATE, update_array[1]
			}
		}'

		PARSE_3='END {
			TOTALS=""
			for (key in UPDATES) {
				TOTALS=TOTALS key "=" UPDATES[key] " "
			}
			printf "%s %s\n", DATE, TOTALS
		}'

		MESSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3"
	fi

elif [ "$KERNEL" = "Darwin" ] ; then
	assertHaveCommand date
	assertHaveCommand softwareupdate

	CMD='eval date ; softwareupdate -l'
	# shellcheck disable=SC2016
	PARSE_0='NR==1 {
		DATE=$0
		PROCESS=0
		TOTAL=0
	}'

	# If the first non-space character is an asterisk, assume this is the name
	# of the update. Otherwise, print the update.
	# shellcheck disable=SC2016
	PARSE_1='NR>1 && PROCESS==1 && $0 !~ /^[[:blank:]]*$/ {
		if ( $0 ~ /^[[:blank:]]*\*/ ) {
			PACKAGE="package=\"" $2 "\""
			RECOMMENDED=""
			RESTART=""
			TOTAL=TOTAL+1
		} else {
			if ( $0 ~ /recommended/ ) { RECOMMENDED="is_recommended=\"true\"" }
			if ( $0 ~ /restart/ ) { RESTART="restart_required=\"true\"" }
			printf "%s %s %s %s\n", DATE, PACKAGE, RECOMMENDED, RESTART
		}
	}'

	# Use sentinel value to skip all text prior to update list.
	# shellcheck disable=SC2016
	PARSE_2='NR>1 && PROCESS==0 && $0 ~ /found[[:blank:]]the[[:blank:]]following/ {
		PROCESS=1
	}'

	PARSE_3='END {
		printf "%s total_updates=%s\n", DATE, TOTAL
	}'

	MESSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3"

else
	# Exits
	failUnsupportedScript
fi

$CMD | tee "$TEE_DEST" | $AWK "$MESSAGE"
echo "Cmd = [$CMD];  | $AWK '$MESSAGE'" >> "$TEE_DEST"
