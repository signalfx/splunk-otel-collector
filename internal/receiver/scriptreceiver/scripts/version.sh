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

PRINTF='END {printf "%s %s %s %s %s %s\n", DATE, MACH_HW_NAME, MACH_ARCH_NAME, OS_REL, OS_NAME, OS_VER}'


if [ "$KERNEL" = "Linux" ] || [ "$KERNEL" = "SunOS" ] || [ "$KERNEL" = "Darwin" ] || [ "$KERNEL" = "FreeBSD" ] ; then
	assertHaveCommand date
	assertHaveCommand uname
	CMD='eval date ; eval uname -m ; eval uname -r ; eval uname -s ; eval uname -v ; eval uname -p'
elif [ "$KERNEL" = "HP-UX" ] ; then
	# HP-UX lacks -p switch.
	assertHaveCommand date
	assertHaveCommand uname
	CMD='eval date ; eval uname -m ; eval uname -r ; eval uname -s ; eval uname -v'
elif [ "$KERNEL" = "AIX" ] ; then
	# AIX uses oslevel for version and release switch.
	assertHaveCommand date
	assertHaveCommand uname
	CMD='eval date ; eval uname -m ; eval oslevel -r ; eval uname -s ; eval oslevel -s'
fi

# Get the date.
# shellcheck disable=SC2016
PARSE_0='NR==1 {DATE=$0}'
# shellcheck disable=SC2016
PARSE_1='NR==2 {MACH_HW_NAME="machine_hardware_name=\"" $0 "\""}'
# shellcheck disable=SC2016
PARSE_2='NR==3 {OS_REL="os_release=\"" $0 "\""}'
# shellcheck disable=SC2016
PARSE_3='NR==4 {OS_NAME="os_name=\"" $0 "\""}'
# shellcheck disable=SC2016
PARSE_4='NR==5 {OS_VER="os_version=\"" $0 "\""}'
# shellcheck disable=SC2016
PARSE_5='NR==6 {MACH_ARCH_NAME="machine_architecture_name=\"" $0 "\""}'

MASSAGE="$PARSE_0 $PARSE_1 $PARSE_2 $PARSE_3 $PARSE_4 $PARSE_5"

$CMD | tee "$TEE_DEST" | $AWK "$MASSAGE $PRINTF"
echo "Cmd = [$CMD];  | $AWK '$MASSAGE $PRINTF'" >> "$TEE_DEST"
