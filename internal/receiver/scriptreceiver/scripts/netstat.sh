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

HEADER='Proto  Recv-Q  Send-Q  LocalAddress                    ForeignAddress                  State'
HEADERIZE="BEGIN {print \"$HEADER\"}"
# shellcheck disable=SC2016
PRINTF='{printf "%-5s  %6s  %6s  %-30.30s  %-30.30s  %-s\n", $1, $2, $3, $4, $5, $6}'
# shellcheck disable=SC2016
FILL_BLANKS='($1=="udp") {$6="<n/a>"}'

if [ "$KERNEL" = "Linux" ] ; then
	queryHaveCommand ss
	FOUND_SS=$?
	if [ $FOUND_SS -eq 0 ] ; then
		CMD='eval ss -antu 2>/dev/null | egrep "tcp|udp"'
		# shellcheck disable=SC2016
		FORMAT='{ state=$2; $2=$3; $3=$4; $4=$5; $5=$6; $6=state}'
	else
		CMD='eval netstat -aenp 2>/dev/null | egrep "tcp|udp"'
	fi
elif [ "$KERNEL" = "SunOS" ] ; then
	CMD='netstat -an -f inet -f inet6'
	FIGURE_SECTION='NR==1 {inUDP=1;inTCP=0} /^TCP: IPv/ {inUDP=0;inTCP=1} /^SCTP:/ {exit}'
	FILTER='/: IPv|Local Address|^$|^-----/ {next}'
	# shellcheck disable=SC2016
	FORMAT_UDP='(inUDP) {localAddr=$1; $1="udp"; $2=$3=0; $4=localAddr; $5="*.*"}'
	# shellcheck disable=SC2016
	FORMAT_TCP='(inTCP) {localAddr=$1; foreignAddr=$2; sendQ=$4; recvQ=$6; state=$7; $1="tcp"; $2=recvQ; $3=sendQ; $4=localAddr; $5=foreignAddr; $6=state}'
	FORMAT="$FORMAT_UDP $FORMAT_TCP"
elif [ "$KERNEL" = "AIX" ] ; then
	CMD='eval netstat -an 2>/dev/null | egrep "tcp|udp"'
elif [ "$KERNEL" = "Darwin" ] ; then
	CMD='eval netstat -anW | egrep "tcp|udp"'
	# shellcheck disable=SC2016
	FORMAT='{gsub("[46]", "", $1)}'
elif [ "$KERNEL" = "HP-UX" ] ; then
    CMD='eval netstat -an | egrep "tcp|udp"'
elif [ "$KERNEL" = "FreeBSD" ] ; then
	# shellcheck disable=SC2089
	CMD='eval netstat -an | egrep "tcp|udp"'
	# shellcheck disable=SC2016
	FORMAT='{gsub("[46]", "", $1)}'
fi

assertHaveCommand "$CMD"
# shellcheck disable=SC2090
$CMD | tee "$TEE_DEST" | $AWK "$HEADERIZE $FIGURE_SECTION $FILTER $FORMAT $FILL_BLANKS $PRINTF"  header="$HEADER"
echo "Cmd = [$CMD];  | $AWK '$HEADERIZE $FIGURE_SECTION $FILTER $FORMAT $FILL_BLANKS $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
