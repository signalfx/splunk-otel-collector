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

CMD='who -H'
HEADER='USERNAME        LINE        HOSTNAME                                  TIME'
# shellcheck disable=SC2016
HEADERIZE='{NR == 1 && $0 = header}'
# shellcheck disable=SC2016
FORMAT='{length(hostname) || hostname=$NF; gsub("[)(]", "",hostname); time=$3; for (i=4; i<=lastTimeColumn; i++) time = time " " $i}'
# shellcheck disable=SC2016
PRINTF='{if (NR == 1) {print $0} else {printf "%-14s  %-10s  %-40.40s  %-s\n", $1,$2,hostname,time}}'

if [ "$KERNEL" = "Linux" ] ; then
	FILL_BLANKS='{hostname = ""; lastTimeColumn = NF-1; if (NF < 5) {hostname = "<console>"; lastTimeColumn = NF}}'
elif [ "$KERNEL" = "SunOS" ] ; then
	FILL_BLANKS='{hostname = ""; lastTimeColumn = NF-1; if (NF < 6) {hostname = "<console>"; lastTimeColumn = NF}}'
elif [ "$KERNEL" = "AIX" ] ; then
	FILL_BLANKS='{hostname = ""; lastTimeColumn = NF-1; if (NF < 6) {hostname = "<console>"; lastTimeColumn = NF}}'
elif [ "$KERNEL" = "HP-UX" ] ; then
    CMD='who -HR'
    FILL_BLANKS='{hostname = ""; lastTimeColumn = NF-1; if (NF < 5) {hostname = "<console>"; lastTimeColumn = NF}}'
elif [ "$KERNEL" = "Darwin" ] ; then
	FILL_BLANKS='{hostname = ""; lastTimeColumn = NF-1; if (NF < 6) {hostname = "<console>"; lastTimeColumn = NF}}'
elif [ "$KERNEL" = "FreeBSD" ] ; then
	FILL_BLANKS='{hostname = ""; lastTimeColumn = NF-1; if (NF < 6) {hostname = "<console>"; lastTimeColumn = NF}}'
fi

assertHaveCommand "$CMD"

out=$($CMD | tee "$TEE_DEST" | $AWK "$HEADERIZE $FILL_BLANKS $FORMAT $PRINTF"  header="$HEADER")
lines=$(echo "$out" | wc -l)
if [ "$lines" -gt 1 ]; then
	echo "$out"
	echo "Cmd = [$CMD];  | $AWK '$HEADERIZE $FILL_BLANKS $FORMAT $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
else
	echo "No data is present" >> "$TEE_DEST"
fi
