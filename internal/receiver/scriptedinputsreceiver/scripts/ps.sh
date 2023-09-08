#!/bin/sh
# Copyright Splunk, Inc.
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

# shellcheck disable=SC2166
if [ "$KERNEL" = "Linux" -o "$KERNEL" = "Darwin" -o "$KERNEL" = "FreeBSD" ] ; then
    assertHaveCommand ps
    CMD='ps auxww'
elif [ "$KERNEL" = "AIX" ] ; then
    assertHaveCommandGivenPath /usr/sysv/bin/ps
    CMD='/usr/sysv/bin/ps -eo user,pid,psr,pcpu,time,pmem,rss,vsz,tty,s,etime,args'
elif [ "$KERNEL" = "SunOS" ] ; then
    assertHaveCommandGivenPath /usr/bin/ps
    CMD='/usr/bin/ps -eo user,pid,psr,pcpu,time,pmem,rss,vsz,tty,s,etime,args'
elif [ "$KERNEL" = "HP-UX" ] ; then
    HEADER='USER               PID   PSR   pctCPU       CPUTIME  pctMEM     RSZ_KB     VSZ_KB   TTY      S       ELAPSED  COMMAND             ARGS'
    # shellcheck disable=SC2016
    FORMAT='{sub("^_", "", $1); if (NF>12) {args=$13; for (j=14; j<=NF; j++) args = args "_" $j} else args="<noArgs>"; sub("^[^\134[: -]*/", "", $12)}'
    # shellcheck disable=SC2016
    PRINTF='{if (NR == 1) {print $0} else {printf "%32.32s  %6s  %4s   %6s  %12s  %6s   %8s   %8s   %-7.7s  %1.1s  %12s  %-100.100s  %s\n", $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, args}}'
    # shellcheck disable=SC2016
    HEADERIZE='{NR == 1 && $0 = header}'

    assertHaveCommand ps
    export UNIX95=1
    CMD='ps -e -o ruser,pid,pset,pcpu,time,vsz,tty,state,etime,args'
    # shellcheck disable=SC2016
    FORMAT='{sub("^_", "", $1); if (NF>12) {args=$13; for (j=14; j<=NF; j++) args = args "_" $j} else args="<noArgs>"; sub("^[\[\]]", "", $11)}'
    # shellcheck disable=SC2016
    PRINTF='{if (NR == 1) {print $0} else {printf "%-14.14s  %6s  %4s   %6s  %12s  %6s   %8s   %8s   %-7.7s  %1.1s  %12s  %-18.18s  %s\n", $1, $2, $3, $4, $5, "?", "?", $6, $7, $8, $9, $10, $11, arg}}'

    $CMD | tee "$TEE_DEST" | $AWK "$HEADERIZE $FORMAT $PRINTF"  header="$HEADER"
    echo "Cmd = [$CMD];  | $AWK '$HEADERIZE $FORMAT $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
    exit
fi

# shellcheck disable=SC2016
# awk logic for adding extra field ARGS with underscore delimiter
ARGS_FORMAT='BEGIN {OFS = "    ";} # specify output field separator
{
    if (NR == 1) # Add extra header/field ARGS in first (header) row
    {
        command_column = NF;
        $(NF+1) = "ARGS";
    }
    else
    {
        # If arguments exist, then append all with underscore delimeter, else specify <noArgs>
        if ($(command_column+1) != "")
        {
            args = $(command_column+1);
            for (i=command_column+2; i<=NF; i++)
            {
                args = args "_" $i;
                $i = "";
            }
            $(command_column+1) = args;
        }
        else
        {
            $(command_column+1) = "<noArgs>";
        }

        # Remove trailing white spaces if any
        sub(/[ \t]+$/,"",$0);
    }
    print;
}'

# Execute the command
$CMD | tee "$TEE_DEST" | $AWK "$ARGS_FORMAT"

echo "Cmd = [$CMD]; $AWK '$ARGS_FORMAT'" >> "$TEE_DEST"
