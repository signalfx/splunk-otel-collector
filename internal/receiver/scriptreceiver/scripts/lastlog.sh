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

HEADER='USERNAME                        FROM                            LATEST'
HEADERIZE="BEGIN {print \"$HEADER\"}"
PRINTF='{printf "%-30s  %-30.30s  %-s\n", username, from, latest}'

if [ "$KERNEL" = "Linux" ] ; then
	CMD='last -iw'
	# shellcheck disable=SC2016
	FILTER='{if ($0 == "") exit; if ($1 ~ /reboot|shutdown/ || $1 in users) next; users[$1]=1}'
	# shellcheck disable=SC2016
	FORMAT='{username = $1; from = (NF==10) ? $3 : "<console>"; latest = $(NF-6) " " $(NF-5) " " $(NF-4) " " $(NF-3)}'
elif [ "$KERNEL" = "SunOS" ] ; then
	CMD='last -n 999'
	# shellcheck disable=SC2016
	FILTER='{if ($0 == "") exit; if ($1 ~ /reboot|shutdown/ || $1 in users) next; users[$1]=1}'
	# shellcheck disable=SC2016
	FORMAT='{username = $1; from = (NF==10) ? $3 : "<console>"; latest = $(NF-6) " " $(NF-5) " " $(NF-4) " " $(NF-3)}'
elif [ "$KERNEL" = "AIX" ] ; then
	failUnsupportedScript
elif [ "$KERNEL" = "Darwin" ] ; then
	CMD='last -99'
	# shellcheck disable=SC2016
	FILTER='{if ($0 == "") exit; if ($1 ~ /reboot|shutdown/ || $1 in users) next; users[$1]=1}'
	# shellcheck disable=SC2016
	FORMAT='{username = $1; from = ($0 !~ /                /) ? $3 : "<console>"; latest = $(NF-6) " " $(NF-5) " " $(NF-4) " " $(NF-3)}'
elif [ "$KERNEL" = "HP-UX" ] ; then
    CMD='lastb -Rx'
	# shellcheck disable=SC2016
    FORMAT='{username = $1; from = ($2=="console") ? $2 : $3; latest = $(NF-3) " " $(NF-2)" " $(NF-1)}'
	# shellcheck disable=SC2016
    FILTER='{if ($1 == "BTMPS_FILE") next; if (NF==0) next; if (NF<=6) next;}'
elif [ "$KERNEL" = "FreeBSD" ] ; then
	CMD='lastlogin'
	# shellcheck disable=SC2016
	FORMAT='{username = $1; from = (NF==8) ? $3 : "<console>"; latest=$(NF-4) " " $(NF-3) " " $(NF-2) " " $(NF-1) " " $NF}'
fi

assertHaveCommand $CMD

out=$($CMD | tee "$TEE_DEST" | $AWK "$HEADERIZE $FILTER $FORMAT $PRINTF"  header="$HEADER")
lines=$(echo "$out" | wc -l)
if [ "$lines" -gt 1 ]; then
	echo "$out"
	echo "Cmd = [$CMD];  | $AWK '$HEADERIZE $FILTER $FORMAT $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
else
	echo "No data is present" >> "$TEE_DEST"
fi
