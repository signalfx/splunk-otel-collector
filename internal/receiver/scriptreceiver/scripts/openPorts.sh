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

# a similar effect can be accomplished with: "nc -z 127.0.0.1 1-32768", and "nc -zu 127.0.0.1 1-32768"

# shellcheck disable=SC1091
. "$(dirname "$0")"/common.sh

HEADER='Proto   Port'
HEADERIZE="BEGIN {print \"$HEADER\"}"
PRINTF='{printf "%-5s  %5d\n", proto, port}'
# shellcheck disable=SC2016
FILTER_INACTIVE='($NF ~ /^CLOSE/) {next}'

if [ "$KERNEL" = "Linux" ] ; then
	queryHaveCommand ss
	FOUND_SS=$?
	if [ $FOUND_SS -eq 0 ] ; then
		CMD='eval ss -lnut | egrep "^tcp|^udp"'
		# shellcheck disable=SC2016
		FORMAT='{proto=$1; sub("^.*:", "", $5); port=$5}'
	else
		CMD='eval netstat -ln | egrep "^tcp|^udp"'
		# shellcheck disable=SC2016
		FORMAT='{proto=$1; sub("^.*:", "", $4); port=$4}'
	fi
elif [ "$KERNEL" = "SunOS" ] ; then
	CMD='netstat -an -f inet -f inet6'
	FIGURE_SECTION='BEGIN {inUDP=1;inTCP=0} /^TCP: IPv/ {inUDP=0;inTCP=1} /^SCTP:/ {exit}'
	FILTER='/: IPv|Local Address|^$|^-----/ {next} (! port) {next}'
	# shellcheck disable=SC2016
	FORMAT='{if (inUDP) proto="udp"; if (inTCP) proto="tcp"; sub("^.*[^0-9]", "", $1); port=$1}'
elif [ "$KERNEL" = "AIX" ] ; then
	CMD='eval netstat -an | egrep "^tcp|^udp"'
	HEADERIZE="BEGIN {print \"$HEADER\"}"
	# shellcheck disable=SC2016
	FORMAT='{gsub("[46]", "", $1); proto=$1; sub("^.*[^0-9]", "", $4); port=$4}'
	# shellcheck disable=SC2016
	FILTER='{if ($4 == "") next}'
elif [ "$KERNEL" = "Darwin" ] ; then
	CMD='eval netstat -ln | egrep "^tcp|^udp"'
	HEADERIZE="BEGIN {print \"$HEADER\"}"
	# shellcheck disable=SC2016
	FORMAT='{gsub("[46]", "", $1); proto=$1; sub("^.*[^0-9]", "", $4); port=$4}'
	# shellcheck disable=SC2016
	FILTER='{if ($4 == "") next}'
elif [ "$KERNEL" = "HP-UX" ] ; then
    CMD='eval netstat -an | egrep "^tcp|^udp"'
    HEADERIZE="BEGIN {print \"$HEADER\"}"
	# shellcheck disable=SC2016
    FORMAT='{gsub("[46]", "", $1); proto=$1; sub("^.*[^0-9]", "", $4); port=$4}'
	# shellcheck disable=SC2016
    FILTER='{if ($4 == "") next}'
elif [ "$KERNEL" = "FreeBSD" ] ; then
# shellcheck disable=SC2089
	CMD='eval netstat -ln | egrep "^tcp|^udp"'
	HEADERIZE="BEGIN {print \"$HEADER\"}"
	# shellcheck disable=SC2016
	FORMAT='{gsub("[46]", "", $1); proto=$1; sub("^.*[^0-9]", "", $4); port=$4}'
fi

assertHaveCommand "$CMD"
# shellcheck disable=SC2090
$CMD | tee "$TEE_DEST" | $AWK "$HEADERIZE $FIGURE_SECTION $FORMAT $FILTER $FILTER_INACTIVE $PRINTF"  header="$HEADER"
echo "Cmd = [$CMD];  | $AWK '$HEADERIZE $FIGURE_SECTION $FORMAT $FILTER $FILTER_INACTIVE $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
