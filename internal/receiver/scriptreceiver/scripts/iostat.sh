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

# suggested command for testing reads: $ find / -type f 2>/dev/null | xargs wc &> /dev/null &

# shellcheck disable=SC1091
. "$(dirname "$0")"/common.sh

if [ "$KERNEL" = "Linux" ] ; then
	CMD='iostat -xky 1 1'
	assertHaveCommand "$CMD"
	# considers the device, r/s and w/s columns and returns output of the first interval
	FILTER='/Device/ && /r\/s/ && /w\/s/ {f=1;}f'
elif [ "$KERNEL" = "SunOS" ] ; then
	CMD='iostat -xn 1 2'
	assertHaveCommand "$CMD"
	# considers the device, r/s and w/s columns and returns output of the second interval
	FILTER='/device/ && /r\/s/ && /w\/s/ {f++;} f==2'
elif [ "$KERNEL" = "AIX" ] ; then
	CMD='iostat  1 2'
	assertHaveCommand "$CMD"
	# considers the disks, kb_read and kb_wrtn columns and returns output of the second interval
	FILTER='/^cd/ {next} /Disks/ && /Kb_read/ && /Kb_wrtn/ {f++;} f==2'
elif [ "$KERNEL" = "FreeBSD" ] ; then
	CMD='iostat -x -c 2'
	assertHaveCommand "$CMD"
	# considers the device, r/s and w/s columns and returns output of the second interval
	FILTER='/device/ && /r\/s/ && /w\/s/ {f++;} f==2'
elif [ "$KERNEL" = "Darwin" ] ; then
	CMD="eval $SPLUNK_HOME/bin/darwin_disk_stats ; sleep 2; echo Pause; $SPLUNK_HOME/bin/darwin_disk_stats"
	# shellcheck disable=SC2086
	assertHaveCommandGivenPath $CMD
	# shellcheck disable=SC2016
	HEADER='Device          rReq_PS      wReq_PS        rKB_PS        wKB_PS  avgWaitMillis   avgSvcMillis   bandwUtilPct'
	HEADERIZE="BEGIN {print \"$HEADER\"}"
	PRINTF='{printf "%-10s  %11s  %11s  %12s  %12s  %13s  %13s  %13s\n", device, rReq_PS, wReq_PS, rKB_PS, wKB_PS, avgWaitMillis, avgSvcMillis, bandwUtilPct}'
	# shellcheck disable=SC2016
	FILTER='BEGIN {FS="|"; after=0} /^Pause$/ {after=1; next} !/Bytes|Operations/ {next} {devices[$1]=$1; values[after,$1,$2]=$3; next}'
	FORMAT='avgSvcMillis=bandwUtilPct="?";'
	FUNC1='function getDeltaPS(disk, metric) {delta=values[1,disk,metric]-values[0,disk,metric]; return delta/2.0}'
	# Calculates the latency by pulling the read and write latency fields from darwin__disk_stats and evaluating their sum
	LATENCY='function getLatency(disk) {read=getDeltaPS(disk,"Latency Time (Read)"); write=getDeltaPS(disk,"Latency Time (Write)"); return expr read + write;}'
	FUNC2='function getAllDeltasPS(disk) {rReq_PS=getDeltaPS(disk,"Operations (Read)"); wReq_PS=getDeltaPS(disk,"Operations (Write)"); rKB_PS=getDeltaPS(disk,"Bytes (Read)")/1024; wKB_PS=getDeltaPS(disk,"Bytes (Write)")/1024; avgWaitMillis=getLatency(disk);}'
	SCRIPT="$HEADERIZE $FILTER $FUNC1 $LATENCY $FUNC2 END {$FORMAT for (device in devices) {getAllDeltasPS(device); $PRINTF}}"
	$CMD | tee "$TEE_DEST" | awk "$SCRIPT"  header="$HEADER"
	echo "Cmd = [$CMD];  | awk '$SCRIPT' header=\"$HEADER\"" >> "$TEE_DEST"
	exit 0
fi

$CMD | tee "$TEE_DEST" | $AWK "$FILTER"
echo "Cmd = [$CMD];  | $AWK '$FILTER'" >> "$TEE_DEST"
