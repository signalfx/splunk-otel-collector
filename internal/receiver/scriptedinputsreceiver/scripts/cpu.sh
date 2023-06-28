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

HEADER='CPU    pctUser    pctNice  pctSystem  pctIowait    pctIdle'
HEADERIZE="BEGIN {print \"$HEADER\"}"
PRINTF='{printf "%-3s  %9s  %9s  %9s  %9s  %9s\n", cpu, pctUser, pctNice, pctSystem, pctIowait, pctIdle}'

if [ "$KERNEL" = "Linux" ] ; then
    queryHaveCommand sar
    FOUND_SAR=$?
    queryHaveCommand mpstat
    FOUND_MPSTAT=$?
    if [ $FOUND_SAR -eq 0 ] ; then
        CMD='sar -P ALL 1 1'
        # shellcheck disable=SC2016
        FORMAT='{cpu=$(NF-6); pctUser=$(NF-5); pctNice=$(NF-4); pctSystem=$(NF-3); pctIowait=$(NF-2); pctIdle=$NF}'
    elif [ $FOUND_MPSTAT -eq 0 ] ; then
        CMD='mpstat -P ALL 1 1'
        # shellcheck disable=SC2016
        FORMAT='{cpu=$(NFIELDS-10); pctUser=$(NFIELDS-9); pctNice=$(NFIELDS-8); pctSystem=$(NFIELDS-7); pctIowait=$(NFIELDS-6); pctIdle=$NF}'
    else
        failLackMultipleCommands sar mpstat
    fi
    # shellcheck disable=SC2016
    FILTER='($0 ~ /CPU/) { if($(NF-1) ~ /gnice/){  NFIELDS=NF; } else {NFIELDS=NF+1;} next} /Average|Linux|^$|%/ {next}'
elif [ "$KERNEL" = "SunOS" ] ; then
    if [ "$SOLARIS_8" = "true" ] || [ "$SOLARIS_9" = "true" ] ; then
        CMD='eval mpstat -a -p 1 2 | tail -1 | sed "s/^[ ]*0/all/"; mpstat -p 1 2 | tail -r'
    else
        CMD='eval mpstat -aq -p 1 2 | tail -1 | sed "s/^[ ]*0/all/"; mpstat -q -p 1 2 | tail -r'
    fi
    assertHaveCommand "$CMD"
    # shellcheck disable=SC2016
    FILTER='($1=="CPU") {exit 1}'
    # shellcheck disable=SC2016
    FORMAT='{cpu=$1; pctUser=$(NF-4); pctNice="0"; pctSystem=$(NF-3); pctIowait=$(NF-2); pctIdle=$(NF-1)}'
elif [ "$KERNEL" = "AIX" ] ; then
    queryHaveCommand mpstat
    queryHaveCommand lparstat
    FOUND_MPSTAT=$?
    FOUND_LPARSTAT=$?
    if [ $FOUND_MPSTAT -eq 0 ] && [ $FOUND_LPARSTAT -eq 0 ] ; then
        # Get extra fields from lparstat
        COUNT=$(lparstat | grep " app" | wc -l)
        if [ $COUNT -gt 0 ] ; then
            # Fetch value from "app" column of lparstat output
            FETCH_APP_COL_NUM='BEGIN {app_col_num = 8}
            {
                if($0 ~ /System configuration|^$/) {next}
                if($0 ~ / app/)
                {
                    for(i=1; i<=NF; i++)
                    {
                        if($i == "app")
                        {
                            app_col_num = i;
                            break;
                        }
                    }
                    print app_col_num;
                    exit 0;
                }
            }'
            APP_COL_NUM=$(lparstat | awk "$FETCH_APP_COL_NUM")
            CPUPool=$(lparstat | tail -1 | awk -v APP_COL_NUM=$APP_COL_NUM -F " " '{print $APP_COL_NUM}')
        else
            CPUPool=0
        fi
        # Fetch other required fields from lparstat output
        OnlineVirtualCPUs=$(lparstat -i | grep "Online Virtual CPUs" | awk -F " " '{print $NF}')
        EntitledCapacity=$(lparstat -i | grep "Entitled Capacity  " | awk -F " " '{print $NF}')
        DEFINE="-v CPUPool=$CPUPool -v OnlineVirtualCPUs=$OnlineVirtualCPUs -v EntitledCapacity=$EntitledCapacity"

        # Get cpu stats using mpstat command and manipulate the output for adding extra fields
        CMD='mpstat -a 1 1'
        # shellcheck disable=SC2016
        FORMAT='BEGIN {flag = 0}
        {
            if($0 ~ /System configuration|^$/) {next}
            if(flag == 1)
            {
                # Prepend extra field values from lparstat
                for(i=NF+4; i>=4; i--)
                {
                    $i = $(i-3);
                }
                if($0 ~ /ALL/)
                {
                    $1 = CPUPool;
                    $2 = OnlineVirtualCPUs;
                    $3 = EntitledCapacity;
                }
                else
                {
                    $1 = "-";
                    $2 = "-";
                    $3 = "-";
                }
            }
            if($0 ~ /cpu /)
            {
                # Prepend extra field headers from lparstat
                for(i=NF+4; i>=4; i--)
                {
                    $i = $(i-3);
                }
                $1 = "CPUPool";
                $2 = "OnlineVirtualCPUs";
                $3 = "EntitledCapacity";
                flag = 1;
            }
            for(i=1; i<=NF; i++)
            {
                printf "%17s ", $i;
            }
            print "";
        }'
    fi
    $CMD | tee "$TEE_DEST" | $AWK $DEFINE "$FORMAT"
    echo "Cmd = [$CMD];  | $AWK $DEFINE '$FORMAT'" >> "$TEE_DEST"
    exit
elif [ "$KERNEL" = "Darwin" ] ; then
    HEADER='CPU    pctUser  pctSystem    pctIdle'
    HEADERIZE="BEGIN {print \"$HEADER\"}"
    PRINTF='{printf "%-3s  %9s  %9s  %9s \n", cpu, pctUser, pctSystem, pctIdle}'
    # top command here is used to get a single instance of cpu metrics
    CMD='top -l 1'
    assertHaveCommand "$CMD"
    # FILTER here skips all the rows that doesn't match "CPU".
    # shellcheck disable=SC2016
    FILTER='($1 !~ "CPU") {next;}'
    # FORMAT here removes '%'in the end of the metrics.
    # shellcheck disable=SC2016
    FORMAT='function remove_char(string, char_to_remove) {
                                    sub(char_to_remove, "", string);
                                    return string;
                            }
                            {
                                cpu="all";
                                pctUser = remove_char($3, "%");
                                pctSystem = remove_char($5, "%");
                                pctIdle = remove_char($7, "%");
                                }'
elif [ "$KERNEL" = "FreeBSD" ] ; then
    CMD='eval top -P -d2 c; top -d2 c'
    assertHaveCommand "$CMD"
    # shellcheck disable=SC2016
    FILTER='($1 !~ "CPU") { next; }'
    # shellcheck disable=SC2016
    FORMAT='function remove_char(string, char_to_remove) {
				sub(char_to_remove, "", string);
				return string;
			}
			{
				if ($1 == "CPU:") {
					cpu = "all";
				} else {
					cpu = remove_char($2, ":");
				}
			}
			{
				pctUser = remove_char($(NF-9), "%");
				pctNice = remove_char($(NF-7), "%");
				pctSystem = remove_char($(NF-5), "%");
				pctIdle = remove_char($(NF-1), "%");
				pctIowait = "0.0";
			}'
elif [ "$KERNEL" = "HP-UX" ] ; then
    queryHaveCommand sar
    FOUND_SAR=$?
    if [ $FOUND_SAR -eq 0 ] ; then
        CMD='sar -M 1 1 ALL'
    fi
    FILTER='/HP-UX|^$|%/ {next}'
    # shellcheck disable=SC2016
    FORMAT='{k=0; if(5<NF) k=1} {cpu=$(1+k); pctUser=$(2+k); pctNice="0"; pctSystem=$(3+k); pctIowait=$(4+k); pctIdle=$(5+k)}'
fi

$CMD | tee "$TEE_DEST" | $AWK "$HEADERIZE $FILTER $FORMAT $PRINTF"  header="$HEADER"
echo "Cmd = [$CMD];  | $AWK '$HEADERIZE $FILTER $FORMAT $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
