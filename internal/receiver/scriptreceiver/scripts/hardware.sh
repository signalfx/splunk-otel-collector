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
# shellcheck disable=SC2016
FORMAT='{key = $1; if (NF == 1) {value = "<notAvailable>"} else {value = $2; for (i=3; i <= NF; i++) value = value " " $i}}'
PRINTF='{printf("%-20s  %-s\n", key, value)}'

if [ "$KERNEL" = "Linux" ] ; then
	assertHaveCommand dmesg
	queryHaveCommand ip
	FOUND_IP=$?
	# CPUs
	CPU_TYPE=$(awk -F: '/model name/ {print $2; exit}' /proc/cpuinfo   2>>"$TEE_DEST")
	CPU_CACHE=$(awk -F: '/cache size/ {print $2; exit}' /proc/cpuinfo  2>>"$TEE_DEST")
	CPU_COUNT=$(grep -c processor /proc/cpuinfo                        2>>"$TEE_DEST")
	# HDs
	# shellcheck disable=SC2010
	for deviceBasename in $(ls /sys/block | grep -E -v '^(dm|md|ram|sr|loop)')
	do
		DEVICE="/sys/block/$deviceBasename"   HARD_DRIVES="$HARD_DRIVES $deviceBasename"
		if [ -e "$DEVICE"/device/model ] ; then HARD_DRIVES="$HARD_DRIVES ($(sed 's/ *$//' "$DEVICE"/device/model))"; fi
		if [ -e "$DEVICE"/size         ] ; then HARD_DRIVES="$HARD_DRIVES $((($(cat "$DEVICE"/size)*512)/(1024*1024*1024))) GB; "; fi
	done
	# NICs
	# For Ubuntu version >= 20, we use cat to read the dmseg file. Otherwise we use dmesg cmd.
	if [ -e "$DMESG_FILE" ] && [ "$UBUNTU_MAJOR_VERSION" -ge 20 ] ; then
		NIC_TYPE=$(cat "$DMESG_FILE"* | awk '/Ethernet/ {sub("[^a-zA-Z]*Ethernet.*$", ""); sub("^[^:]*: ", ""); print; exit}')
	else
		NIC_TYPE=$(dmesg | awk '/Ethernet/ {sub("[^a-zA-Z]*Ethernet.*$", ""); sub("^[^:]*: ", ""); print; exit}')
	fi
	if [ $FOUND_IP -eq 0 ]; then
		NIC_COUNT=$(ip a | awk  '!length() || $2 ~/lo/ || /^ / {next} {ct++} END {print ct}')
	else
		assertHaveCommand ifconfig
		NIC_COUNT=$(ifconfig | awk '!length() || /^( |lo)/ {next} {ct++} END {print ct}')
	fi
	# memory
	MEMORY_REAL=$(awk -F: '/MemTotal/ {print $2; exit}' /proc/meminfo  2>>"$TEE_DEST")
	MEMORY_SWAP=$(awk -F: '/SwapTotal/ {print $2; exit}' /proc/meminfo 2>>"$TEE_DEST")
elif [ "$KERNEL" = "SunOS" ] ; then
	UNAME_PLATFORM=$(uname -i)
	assertHaveCommand mpstat
	assertHaveCommand iostat
	assertHaveCommand dmesg
	assertHaveCommandGivenPath /usr/sbin/prtconf
	assertHaveCommandGivenPath /usr/sbin/swap
	# CPUs and NIC count
	if [ -x /usr/sbin/prtdiag ] ; then
		if [ "$SOLARIS_10" = "true" ] || [ "$SOLARIS_11" = "true" ] ; then
			# shellcheck disable=SC2016
			CPU_TYPE=$(/usr/sbin/prtdiag | $AWK 'BEGIN {leftToSkip=-1} /Processor Sockets/ {leftToSkip=3; next} (leftToSkip>0) {leftToSkip-=1; next} (!leftToSkip) {sub("[0-9]$", "", $0); sub(" CPU socket #$", "", $0); print $0; exit}')
		else
			# shellcheck disable=SC2016
			CPU_TYPE=$(/usr/sbin/prtdiag | $AWK 'BEGIN {leftToSkip=-1} /Processor Sockets/ {leftToSkip=3; next} (leftToSkip>0) {leftToSkip-=1; next} (!leftToSkip) {sub("[0-9]$", "", $0); sub(" [A-Za-z]+ ?$", "", $0); print $0; exit}')
		fi
		NIC_COUNT=$(/usr/sbin/prtdiag | grep -c NIC)
	elif [ -x /usr/platform/"$UNAME_PLATFORM"/sbin/prtdiag ]; then
		# shellcheck disable=SC2016
		CPU_TYPE=$(/usr/platform/"$UNAME_PLATFORM"/sbin/prtdiag | $AWK 'BEGIN {leftToSkip=-1} /Processor Sockets/ {leftToSkip=3; next} (leftToSkip>0) {leftToSkip-=1; next} (!leftToSkip) {sub("[0-9]$", "", $0); sub(" [A-Za-z]+ ?$", "", $0); print $0; exit}')
		NIC_COUNT=$(/usr/platform/"$UNAME_PLATFORM"/sbin/prtdiag | grep -c NIC)
	else
		echo "Not found commandGivenPath [ /usr/sbin/prtdiag or /usr/platform/$UNAME_PLATFORM/sbin/prtdiag ] on this host, quitting" >> "$TEE_DEST"
		exit 1
	fi
	# shellcheck disable=SC2016
	CPU_CACHE=$(/usr/sbin/prtconf -v | $AWK 'function hexToDecKB (hex, digitsAll, idx, curDigit, dec) {sub("^value=", "", hex); for (idx=1; idx<=length(hex); idx++) {curDigit = index("0123456789abcdef", substr(hex,idx,1)); dec=(16*dec)+curDigit-1} if (debug) printf "hexToDec:%s->%d ", hex, dec; dec /= 1024; return dec} BEGIN {L2=L1i=L1d=0} (L2) {strL2=$1; L2=0} /l2-cache-size/ {L2=1} (L1i) {strL1i=$1; L1i=0} /l1-icache-size/ {L1i=1} (L1d) {strL1d=$1; L1d=0} /l1-dcache-size/ {L1d=1} END {if (debug) printf "strL2:%s strL1i:%s strL1d:%s ", strL2, strL1i, strL1d; nL2=hexToDecKB(strL2); nL1=hexToDecKB(strL1i)+hexToDecKB(strL1d); printf "L1:%dKB L2:%dKB", nL1, nL2}' debug="$DEBUG")
	if [ "$SOLARIS_8" = "true" ] || [ "$SOLARIS_9" = "true" ] ; then
		CPU_COUNT=$(mpstat    | grep -cv CPU)
	else
		CPU_COUNT=$(mpstat -q | grep -cv CPU)
	fi
	# # # that gives # of cores; `/usr/sbin/psrinfo -p` gives # of chips
	# HDs
	# shellcheck disable=SC2016
	HARD_DRIVES=$(iostat -E | $AWK '/Soft Errors:/ {name=$1} /^Vendor:/ {info = $2 " " $4} /^Size:/ {sizeGB=0+$2; if (sizeGB>0) drives[name]=info " " $2} END {for (d in drives) printf("%s %s; ", d, drives[d])}')
	# NICs
	NIC_TYPE=$(dmesg | grep 'mac address' | sed -n 's/^.*] [a-z]*[0-9]*: //;s/mac address .*$//;p' | uniq)
	# memory
	MEMORY_REAL=$(/usr/sbin/prtconf | awk '/^Memory size:/ {print $3 " MB"; exit}')
	# shellcheck disable=SC2016
	MEMORY_SWAP=$(/usr/sbin/swap -s | $AWK '{used=0+$(NF-3); free=0+$(NF-1); total=(used+free)/1024; print int(total) " MB"}')
elif [ "$KERNEL" = "AIX" ] ; then
	assertHaveCommandGivenPath /usr/sbin/prtconf
	assertHaveCommandGivenPath /usr/sbin/lsattr
	assertHaveCommandGivenPath /usr/sbin/lsdev
	assertHaveCommandGivenPath /usr/sbin/lscfg
	assertHaveCommandGivenPath /usr/sbin/lspv
	assertHaveCommandGivenPath /usr/sbin/lsps
	# CPUs
	# shellcheck disable=SC2016
	CPU_TYPE=$(/usr/sbin/prtconf | $AWK -F: '/^Processor Type:/{type=$2} /^Processor Clock Speed:/ {clock=$2}END {printf("%s %s",type,clock)}')
	# shellcheck disable=SC2016
	CPU_CACHE=$(/usr/sbin/lsattr -EHl L2cache0 | $AWK '/^size/{print "L2:" $2 " KB" }')
	CPU_COUNT=$(/usr/sbin/lsdev -Cc processor | grep -c proc)
	# HDs
        HDD_NAME=$(/usr/sbin/lsdev -Cc disk | awk '{print $1}')
        HARD_DRIVES=""
        for disk in $HDD_NAME
        do
		# shellcheck disable=SC2016
		HARD_INFO=$(/usr/sbin/lscfg -vpl "$disk" | $AWK -F . '/Manufacturer/ {name = $NF } /Machine Type and Model/ {info = $(NF)} END {printf("%s %s", name, info)}')
		ACTIVE_STATUS=$(/usr/sbin/lspv | awk -v pat="$disk" '$0~pat{print $NF}')
		VOLUME_GROUP=$(/usr/sbin/lspv | awk -v pat="$disk" '$0~pat{print $3}')

		if [ "${ACTIVE_STATUS}" != "active" ] || [ "${VOLUME_GROUP}" = "None" ]; then # lspv cannot get disk-size as disk is inactive or not in any volume group
			HARD_MB=$(getconf DISK_SIZE /dev/"$disk")" MB"
		else
			HARD_MB=$(/usr/sbin/lspv -L "$disk" | awk -F \( '{print $2}'| awk '/VG DESCRIPTORS/{print $1" MB"}')
		fi
                HARD_DRIVES="$HARD_DRIVES$disk $HARD_INFO $HARD_MB; "
        done
	# NICs
	NIC_TYPE=$(/usr/sbin/lsdev -Cc adapter | grep ent | awk -F"    " '{print $1" "$3"; "}')
	NIC_COUNT=$(/usr/sbin/lsdev -Cc adapter | grep -c ent)
	# memory
	# shellcheck disable=SC2016
	MEMORY_REAL=$(/usr/sbin/lsattr -EHl mem0 | $AWK '/^size/ {print $2 " MB"}')
	# shellcheck disable=SC2016
	MEMORY_SWAP=$(/usr/sbin/lsps -s | $AWK -F MB '/MB/ {print $1" MB"}')
elif [ "$KERNEL" = "Darwin" ] ; then
	assertHaveCommand sysctl
	assertHaveCommand df
	assertHaveCommand system_profiler
	assertHaveCommand ifconfig
	# CPUs
	CPU_TYPE=$(sysctl machdep.cpu.brand_string | sed -E 's/^.*: //;s/[ ]+/ /g')
	CPU_CACHE=$(sysctl hw.cachesize | awk '{L1=$3/1024; L2=$4/(1024*1024); printf "L1:%d KB; L2:%d MB", L1, L2}')
	CPU_COUNT=$(sysctl hw.ncpu | sed 's/^.*: //')
	# HDs
	HARD_DRIVES=$(df -h | awk '/^\/dev/ {sub("^.*\134/", "", $1); drives[$1] = $2} END {for(d in drives) printf("%s: %s; ", d, drives[d])}')
	# NICs
	NIC_TYPE=$(system_profiler SPNetworkDataType | awk '/Media Subtype:/ {print $3; exit}')
	NIC_COUNT=$(ifconfig | grep -c 'supported media:.*baseT')
	# memory
	MEMORY_REAL=$(sysctl hw.memsize | awk '{print $2/(1024*1024) " MB"}')
	MEMORY_SWAP=$(sysctl vm.swapusage | awk '{print 0+$4 " MB"}')
elif [ "$KERNEL" = "HP-UX" ] ; then
    assertHaveCommand ioscan
    assertHaveCommand iostat
    assertHaveCommand lanscan
    assertHaveCommand machinfo
    assertHaveCommand swapinfo
    OUTPUT=$(machinfo)
    CPU_TYPE=$(echo "$OUTPUT" | awk '/processor family/ { for(i=4; i<=NF; i++) printf("%s ", $i); exit}')
    CPU_CACHE=$(echo "$OUTPUT" | awk '/L[123]/ {cache+=$5} END {print cache " KB"}')
    CPU_COUNT=$(echo "$OUTPUT" | awk '/CPUs/ {print $5; exit}')
    HARD_DRIVES=$(iostat 2 1 | wc -l)
	# shellcheck disable=SC2307,2003
    HARD_DRIVES=$(expr "$HARD_DRIVES"-4)
    NIC_COUNT=$(lanscan -i | wc -l)
    NIC_TYPE=$(ioscan -u | grep lan | awk 'NF>2 {for(i=3; i<=NF; i++) printf("%s", $i); exit}')
    OUTPUT=$(swapinfo -tm)
    MEMORY_REAL=$(echo "$OUTPUT" | awk '$1=="memory" {print $2 " MB"; exit}')
    MEMORY_SWAP=$(echo "$OUTPUT" | awk '$1=="dev" {print $2 " MB"; exit}')
elif [ "$KERNEL" = "FreeBSD" ] ; then
	assertHaveCommand sysctl
	assertHaveCommand df
	assertHaveCommand ifconfig
	assertHaveCommand dmesg
	assertHaveCommand top
	# CPUs
	CPU_TYPE=$(sysctl hw.model | sed 's/^.*: //')
	CPU_CACHE=
	CPU_COUNT=$(sysctl hw.ncpu | sed 's/^.*: //')
	# HDs
	HARD_DRIVES=$(df -h | awk '/^\/dev/ {sub("^.*\134/", "", $1); drives[$1] = $2} END {for(d in drives) printf("%s: %s; ", d, drives[d])}')
	# NICs
	IFACE_NAME=$(ifconfig -a | awk '!/^[a-z]/ {next} /LOOPBACK/ {next} {print $1}' | head -1)
	NIC_TYPE=$(dmesg | awk '(index($0, iface) && index($0, " port ")) {sub("^.*<", ""); sub(">.*$", ""); print $0}' iface="$IFACE_NAME" | head -1)
	NIC_COUNT=$(ifconfig -a | grep -c media)
	# memory
	MEMORY_REAL=$(sysctl hw.physmem | awk '{print $2/(1024*1024) "MB"}')
	MEMORY_SWAP=$(top -Sb 0 | awk '/^Swap: / {print $2 "B"}')
fi

formatAndPrint ()
{
	# shellcheck disable=SC2086
	echo $1 | awk "$FORMAT $PRINTF"
}

formatAndPrint "KEY           VALUE"
formatAndPrint "CPU_TYPE      $CPU_TYPE"
formatAndPrint "CPU_CACHE     $CPU_CACHE"
formatAndPrint "CPU_COUNT     $CPU_COUNT"
formatAndPrint "HARD_DRIVES   $HARD_DRIVES"
formatAndPrint "NIC_TYPE      $NIC_TYPE"
formatAndPrint "NIC_COUNT     $NIC_COUNT"
formatAndPrint "MEMORY_REAL   $MEMORY_REAL"
formatAndPrint "MEMORY_SWAP   $MEMORY_SWAP"
