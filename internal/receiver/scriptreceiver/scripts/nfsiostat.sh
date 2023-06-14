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

HEADER='Mount 		  Path		  r_op/s    w_op/s    r_KB/s    w_KB/s    rpc_backlog    r_avg_RTT    w_avg_RTT    r_avg_exe    w_avg_exe'
HEADERIZE="BEGIN {print \"$HEADER\"}"

# We can have the multiple mounts for the nfs. So we have to parse mount separately.
# For CentOS and RHEL the number of lines for each mount is 9, while for the ubuntu it is 22
# due to the bug mentioned in this link. https://bugs.launchpad.net/ubuntu/+source/nfs-utils/+bug/1584719
# So, we are handling the case of Ubuntu separately.

# When awk iterates through each line, using modulo operator we are checking the line number
# And extracting the particular value from that line and assigning it to the variable
# which we will use when the output of modulo is 0 as it will be the last line of that mount.

# We are also removing last character in the line "path=substr($4, 1, length($4)-1)"
# as last character of the path is ":"

if [ "$KERNEL" = "Linux" ] ; then
    CMD='nfsiostat'

    assertHaveCommand $CMD
    no_of_lines=$($CMD| wc -l)

    # If there are no mount, exit
    if [ "$no_of_lines" -eq 1 ];
    then
        $CMD >> "$TEE_DEST"
        exit 1
    fi
    OS_FILE=/etc/os-release
    # Below condition is added to handle the case of Ubuntu OS
    if [ -e $OS_FILE ] && (awk -F'=' '/ID=/ {print $2}' $OS_FILE | grep -q ubuntu);
    then
        # shellcheck disable=SC2016
        OS_RELEASE=$($AWK -F= '/VERSION_ID=/ {print $2}' $OS_FILE)
        if [ "$OS_RELEASE" = "\"18.04\"" ] || [ "$OS_RELEASE" = "\"20.04\"" ] || [ "$OS_RELEASE" = "\"22.04\"" ] ; then # Ubuntu 18.04, 20.04 and 22.04
            # shellcheck disable=SC2016
            FORMAT='{
                if (NR%10==2){
                echo "device"
                    device=$1
                    path=substr($4, 1, length($4)-1)
                }
                else if (NR%10==5){
                    rpc_backlog=$2
                }
                else if (NR%10==8){
                    r_op_s=$1
                    r_kb_s=$2
                    r_avg_rtt=$6
                    r_avg_exe=$7
                }
                else if (NR%10==0){
                    w_op_s=$1
                    w_kb_s=$2
                    w_avg_rtt=$6
                    w_avg_exe=$7
                    printf "%s %s %s %s %s %s %s %s %s %s %s\n",device, path, r_op_s, w_op_s, r_kb_s, w_kb_s, rpc_backlog, r_avg_rtt, w_avg_rtt, r_avg_exe, w_avg_exe
                }
            }'
        else
            # shellcheck disable=SC2016
            FORMAT='{
                if (NR%22==2){
                echo "device"
                    device=$1
                    path=substr($4, 1, length($4)-1)
                }
                else if (NR%22==6){
                    rpc_backlog=$1
                }
                else if (NR%22==9){
                    r_op_s=$1
                }
                else if (NR%22==10){
                    r_kb_s=$1
                }
                else if (NR%22==13){
                    r_avg_rtt=$1
                }
                else if (NR%22==14){
                    r_avg_exe=$1
                }
                else if (NR%22==17){
                    w_op_s=$1
                }
                else if (NR%22==18){
                    w_kb_s=$1
                }
                else if (NR%22==21){
                    w_avg_rtt=$1
                }
                else if (NR%22==0){
                    w_avg_exe=$1
                    printf "%s %s %s %s %s %s %s %s %s %s %s\n",device, path, r_op_s, w_op_s, r_kb_s, w_kb_s, rpc_backlog, r_avg_rtt, w_avg_rtt, r_avg_exe, w_avg_exe
                }
            }'
        fi
    # For CentOS and RHEL
    else
        #For  RHEL 8.x
        if [ -e $OS_FILE ] && ( ( (awk -F'=' '/ID=/ {print $2}' $OS_FILE | grep -q rhel) && (awk -F'=' '/VERSION_ID=/ {print $2}' $OS_FILE | grep -Eq 8.7\|8.6\|8.5\|8.4\|8.3\|9) ) || ( (awk -F'=' '/ID=/ {print $2}' $OS_FILE | grep -q cent) && (awk -F'=' '/VERSION_ID=/ {print $2}' $OS_FILE | grep -Eq 8) ) );
        then
            # shellcheck disable=SC2016
            FORMAT='{
                if (NR%10==2){
                    device=$1
                    path=substr($4, 1, length($4)-1)
                }
                else if (NR%10==5){
                    rpc_backlog=$2
                }
                else if (NR%10==8){
                    r_op_s=$1
                    r_kb_s=$2
                    r_avg_rtt=$6
                    r_avg_exe=$7
                }
                else if (NR%10==0){
                    w_op_s=$1
                    w_kb_s=$2
                    w_avg_rtt=$6
                    w_avg_exe=$7
                    printf "%s %s %s %s %s %s %s %s %s %s %s\n",device, path, r_op_s, w_op_s, r_kb_s, w_kb_s, rpc_backlog, r_avg_rtt, w_avg_rtt, r_avg_exe, w_avg_exe
                }
            }'
        else
            # shellcheck disable=SC2016
            FORMAT='{
                if (NR%9==2){
                    device=$1
                    path=substr($4, 1, length($4)-1)
                }
                else if (NR%9==5){
                    rpc_backlog=$2
                }
                else if (NR%9==7){
                    r_op_s=$1
                    r_kb_s=$2
                    r_avg_rtt=$6
                    r_avg_exe=$7
                }
                else if (NR%9==0){
                    w_op_s=$1
                    w_kb_s=$2
                    w_avg_rtt=$6
                    w_avg_exe=$7
                    printf "%s %s %s %s %s %s %s %s %s %s %s\n",device, path, r_op_s, w_op_s, r_kb_s, w_kb_s, rpc_backlog, r_avg_rtt, w_avg_rtt, r_avg_exe, w_avg_exe
                }
            }'
        fi
    fi
    $CMD | tee "$TEE_DEST" |  awk "$HEADERIZE $FORMAT" | column -t
    echo "Cmd = [$CMD];  | awk '$HEADERIZE $FORMAT' header=\"$HEADER\"" >> "$TEE_DEST"

else
    failUnsupportedScript
fi
