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

# jscpd:ignore-start
# shellcheck disable=SC1091
. "$(dirname "$0")"/common.sh

HEADER='Name       MAC                inetAddr         inet6Addr                                  Collisions  RXbytes          RXerrors         TXbytes          TXerrors         Speed        Duplex'
FORMAT='{mac = length(mac) ? mac : "?"; collisions = length(collisions) ? collisions : "?"; RXbytes = length(RXbytes) ? RXbytes : "?"; RXerrors = length(RXerrors) ? RXerrors : "?"; TXbytes = length(TXbytes) ? TXbytes : "?"; TXerrors = length(TXerrors) ? TXerrors : "?"; speed = length(speed) ? speed : "?"; duplex = length(duplex) ? duplex : "?"}'
PRINTF='END {printf "%-10s %-17s  %-15s  %-42s %-10s  %-16s %-16s %-16s %-16s %-12s %-12s\n", name, mac, IPv4, IPv6, collisions, RXbytes, RXerrors, TXbytes, TXerrors, speed, duplex}'

if [ "$KERNEL" = "Linux" ] ; then
	HEADER='Name       MAC                inetAddr         inet6Addr                                  Collisions  RXbytes          RXerrors         RXdropped          TXbytes          TXerrors         TXdropped          Speed        Duplex'
	PRINTF='END {printf "%-10s %-17s  %-15s  %-42s %-10s  %-16s %-16s %-18s %-16s %-16s %-18s %-12s %-12s\n", name, mac, IPv4, IPv6, collisions, RXbytes, RXerrors, RXdropped, TXbytes, TXerrors, TXdropped, speed, duplex}'
	queryHaveCommand ip
	FOUND_IP=$?
	if [ $FOUND_IP -eq 0 ]; then
		CMD_LIST_INTERFACES="eval ip -s a | tee $TEE_DEST|grep 'state UP' | grep mtu | grep -Ev lo | tee -a $TEE_DEST | cut -d':' -f2 | tee -a $TEE_DEST | cut -d '@' -f 1 | tee -a $TEE_DEST | sort -u | tee -a $TEE_DEST"
		# shellcheck disable=SC2016
		CMD='eval ip addr show $iface; ip -s link show'
		# shellcheck disable=SC2016
		GET_IPv4='{if ($0 ~ /inet /) {split($2, a, " "); IPv4 = a[1]}}'
		# shellcheck disable=SC2016
		GET_IPv6='{if ($0 ~ /inet6 /) { IPv6 = $2 }}'
		# shellcheck disable=SC2016
		GET_TXbytes='{
			if($0 ~ /TX: /){
				tx_row_count=NR+1;
				for(i=1;i<=NF;i++){
					if($i=="bytes"){
						TX_bytes_column=i;
					}
					else if($i=="errors"){
						TX_errors_column=i;
					}
					else if($i=="dropped"){
						TX_dropped_column=i;
					}
					else if($i=="collsns"){
						TX_collsns_column=i;
					}
				}
				next;
			}
			if(NR==tx_row_count){
				(TX_bytes_column == "") ? TXbytes = 0  : TXbytes = $(TX_bytes_column - 1);
				(TX_errors_column == "") ? TXerrors = "<n/a>"  : TXerrors = $(TX_errors_column - 1);
				(TX_dropped_column == "") ? TXdropped = "<n/a>"  : TXdropped = $(TX_dropped_column - 1);
				(TX_collsns_column == "") ? collisions = 0  : collisions = $(TX_collsns_column - 1);
			}
		}'
		# shellcheck disable=SC2016
		GET_RXbytes='{
			if($0 ~ /RX: /){
				rx_row_count=NR+1;
				for(i=1;i<=NF;i++){
					if($i=="bytes"){
						RX_bytes_column=i;
					}
					else if($i=="errors"){
						RX_errors_column=i;
					}
					else if($i=="dropped"){
						RX_dropped_column=i;
					}
				}next;
			}
			if(NR==rx_row_count){
				(RX_bytes_column == "") ? RXbytes = 0  : RXbytes = $(RX_bytes_column - 1);
				(RX_errors_column == "") ? RXerrors = "<n/a>"  : RXerrors = $(RX_errors_column - 1);
				(RX_dropped_column == "") ? RXdropped = "<n/a>"  : RXdropped = $(RX_dropped_column - 1);
			}
		}'
	else
		assertHaveCommand ifconfig
		# shellcheck disable=SC2089
		CMD_LIST_INTERFACES="eval ifconfig | tee $TEE_DEST | grep 'Link encap:\|mtu' | grep -Ev lo | tee -a $TEE_DEST | cut -d' ' -f1 | cut -d':' -f1 | tee -a $TEE_DEST | sort -u | tee -a $TEE_DEST"
		CMD='ifconfig'
		# shellcheck disable=SC2016
		GET_IPv4='{if ($0 ~ /inet addr:/) {split($2, a, ":"); IPv4 = a[2]} else if ($0 ~ /inet /) {IPv4 = $2}}'
		# shellcheck disable=SC2016
		GET_IPv6='{if ($0 ~ /inet6 addr:/) { IPv6 = $3 } else if ($0 ~ /inet6 /) { IPv6 = $2 }}'
		# shellcheck disable=SC2016
		GET_COLLISIONS='{
			if ($0 ~ /collisions:/){
				for(i=1;i<=NF;i++){
					if($i  ~ /collisions:/){
						collisions_col_no = i;
						break;
					}
				}
				if(collisions_col_no==""){
					collisions=0;
				}
				else
					split($collisions_col_no, a, ":");
					collisions=a[2];
			}
			else if($0 ~ /collisions /){
				for(i=1;i<=NF;i++){
					if($i=="collisions"){
						collisions_column=i+1;
					}
				}
				(collisions_column != "") ? collisions = $collisions_column : collisions = 0;
			}
		}'
		# shellcheck disable=SC2016
		GET_RXbytes='{
			if ($0 ~ /RX bytes:/){
				for(i=1;i<=NF;i++){
					if($i  ~ /bytes:/){
						rxbytes_col_no = i;
						break;
					}
				}
				if(rxbytes_col_no==""){
					RXbytes=0;
				}
				else
					split($rxbytes_col_no, a, ":");
					RXbytes=a[2];
			}
			else if($0 ~ /RX/ && $0 ~ /bytes/){
				for(i=1;i<=NF;i++){
					if($i=="bytes"){
						RXbytes_column=i+1;
						row = NR;
					}
				}
				if(NR == row){
					if(RXbytes_column != ""){
						RXbytes = $RXbytes_column;
					}
					else
						RXbytes = 0;
				}
			}
		}'
		# shellcheck disable=SC2016
		GET_RXerrors='{
			if ($0 ~ /RX packets:/){
				for(i=1;i<=NF;i++){
					if($i  ~ /errors:/){
						rxerrors_col_no = i;
					}
					else if($i  ~ /dropped:/){
						rxdropped_col_no = i;
					}
				}
				if(rxerrors_col_no != ""){
					split($rxerrors_col_no, a, ":");
					RXerrors=a[2];
				}
				else
					RXerrors="<n/a>";
				if(rxdropped_col_no != ""){
					split($rxdropped_col_no, b, ":");
					RXdropped=b[2];
				}
				else
					RXdropped="<n/a>";
			}
			else if($0 ~ /RX/ && ($0 ~ /errors/)){
				for(i=1;i<=NF;i++){
					if($i=="errors"){
						RXerrors_column=i+1;
					}
					if($i=="dropped"){
						RXdropped_column=i+1;
					}
				}
				(RXerrors_column != "") ? RXerrors=$RXerrors_column : RXerrors = "<n/a>";
				(RXdropped_column != "") ? RXdropped = $RXdropped_column : RXdropped = "<n/a>";
			}
		}'
		# shellcheck disable=SC2016
		GET_TXbytes='{
			if ($0 ~ /TX bytes:/){
				for(i=1;i<=NF;i++){
					if($i  ~ /bytes:/){
						txbytes_col_no = i;
					}
				}
				if(txbytes_col_no==""){
					TXbytes=0;
				}
				else
					split($txbytes_col_no, a, ":");
					TXbytes=a[2];
			}
			else if($0 ~ /TX/ && $0 ~ /bytes/){
				for(i=1;i<=NF;i++){
					if($i=="bytes"){
						TXbytes_column=i+1;
						row = NR;
					}
				}
				if(NR == row){
					if(TXbytes_column != ""){
						TXbytes = $TXbytes_column;
					}
					else
						TXbytes = 0;
				}
			}
		}'
		# shellcheck disable=SC2016
		GET_TXerrors='{
			if ($0 ~ /TX packets:/){
				for(i=1;i<=NF;i++){
					if($i  ~ /errors:/){
						txerrors_col_no = i;
					}
					if($i  ~ /dropped:/){
						txdropped_col_no = i;
					}
				}
				if(txerrors_col_no != ""){
					split($txerrors_col_no, a, ":");
					TXerrors=a[2];
				}
				else
					TXerrors="<n/a>";
				if(txdropped_col_no != ""){
					split($txdropped_col_no, b, ":");
					TXdropped=b[2];
				}
				else
					TXdropped="<n/a>";
			}
			else if($0 ~ /TX/ && $0 ~ /errors/){
				for(i=1;i<=NF;i++){
					if($i=="errors"){
						TXerrors_column=i+1;
					}
					if($i=="dropped"){
						TXdropped_column=i+1;
					}
				}
				(TXerrors_column != "") ? TXerrors = $TXerrors_column : TXerrors = "<n/a>";
				(TXdropped_column != "") ? TXdropped = $TXdropped_column : TXdropped = "<n/a>";
			}
		}'
	fi
	GET_ALL="$GET_IPv4 $GET_IPv6 $GET_COLLISIONS $GET_RXbytes $GET_RXerrors $GET_TXbytes $GET_TXerrors"
	FILL_BLANKS='{length(speed) || speed = "<n/a>"; length(duplex) || duplex = "<n/a>"; length(TXdropped) || TXdropped = "<n/a>";length(RXdropped) || RXdropped = "<n/a>"; length(IPv4) || IPv4 = "<n/a>"; length(IPv6) || IPv6= "<n/a>"}'
	BEGIN='BEGIN {RXbytes = TXbytes = collisions = 0}'
	# shellcheck disable=SC2090
	out=$($CMD_LIST_INTERFACES)
	lines=$(echo "$out" | wc -l)
	if [ "$lines" -gt 0 ]; then
		echo "$HEADER"
	fi
	for iface in $out
	do
		if [ -r /sys/class/net/"$iface"/duplex ]; then
			DUPLEX=$(cat /sys/class/net/"$iface"/duplex 2>/dev/null || echo 'error')
			if [ "$DUPLEX" != 'error' ]; then
				DUPLEX=$(echo "$DUPLEX" | sed 's/./\u&/')
				if [ -r /sys/class/net/"$iface"/speed ]; then
					SPEED=$(cat /sys/class/net/"$iface"/speed 2>/dev/null || echo 'error')
					[ -n "$SPEED" ] && [ "$SPEED" != 'error' ] && SPEED="${SPEED}Mb/s"
				else
					# For Ubuntu version >= 20, we use cat to read the dmseg file. Otherwise we use dmesg cmd.
					if [ -e "$DMESG_FILE" ] && [ "$UBUNTU_MAJOR_VERSION" -ge 20 ] ; then
						SPEED=$(cat "$DMESG_FILE"*  | awk '/[Ll]ink( is | )[Uu]p/ && /'"$iface"'/ {for (i=1; i<=NF; ++i) {if (match($i, /([0-9]+)([Mm]bps)/))             {print $i} else { if (match($i, /[Mm]bps/))   {print $(i-1) "Mb/s"} } } }' | sed '$!d')
					else
						assertHaveCommand dmesg
						SPEED=$(dmesg  | awk '/[Ll]ink( is | )[Uu]p/ && /'"$iface"'/ {for (i=1; i<=NF; ++i) {if (match($i, /([0-9]+)([Mm]bps)/))             {print $i} else { if (match($i, /[Mm]bps/))   {print $(i-1) "Mb/s"} } } }' | sed '$!d')
					fi
				fi
			else
				DUPLEX=""
			fi
		fi
		if [ "$DUPLEX" = "" ] || [ "$SPEED" = "" ] ; then
			assertHaveCommand dmesg
			# Get Duplex only if still null
			if [ "$DUPLEX" = "" ] ; then
				# For Ubuntu version >= 20, we use cat to read the dmseg file. Otherwise we use dmesg cmd.
				if [ -e "$DMESG_FILE" ] && [ "$UBUNTU_MAJOR_VERSION" -ge 20 ] ; then
					DUPLEX=$(cat "$DMESG_FILE"* | awk '/[Ll]ink( is | )[Uu]p/ && /'"$iface"'/ {for (i=1; i<=NF; ++i) {if (match($i, /([-_a-zA-Z0-9]+)([Dd]uplex)/)) {print $i} else { if (match($i, /[Dd]uplex/)) {print $(i-1)       } } } }' | sed 's/[-_]//g; $!d')
				else
					DUPLEX=$(dmesg | awk '/[Ll]ink( is | )[Uu]p/ && /'"$iface"'/ {for (i=1; i<=NF; ++i) {if (match($i, /([-_a-zA-Z0-9]+)([Dd]uplex)/)) {print $i} else { if (match($i, /[Dd]uplex/)) {print $(i-1)       } } } }' | sed 's/[-_]//g; $!d')
				fi
			fi
			# Get Speed only if still null
			if [ "$SPEED" = "" ] ; then
				# For Ubuntu version >= 20, we use cat to read the dmseg file. Otherwise we use dmesg cmd.
				if [ -e "$DMESG_FILE" ] && [ "$UBUNTU_MAJOR_VERSION" -ge 20 ] ; then
					SPEED=$(cat "$DMESG_FILE"*  | awk '/[Ll]ink( is | )[Uu]p/ && /'"$iface"'/ {for (i=1; i<=NF; ++i) {if (match($i, /([0-9]+)([Mm]bps)/))             {print $i} else { if (match($i, /[Mm]bps/))   {print $(i-1) "Mb/s"} } } }' | sed '$!d')
				else
					SPEED=$(dmesg  | awk '/[Ll]ink( is | )[Uu]p/ && /'"$iface"'/ {for (i=1; i<=NF; ++i) {if (match($i, /([0-9]+)([Mm]bps)/))             {print $i} else { if (match($i, /[Mm]bps/))   {print $(i-1) "Mb/s"} } } }' | sed '$!d')
				fi
			fi
		fi
		if [ $FOUND_IP -eq 0 ]; then
			# shellcheck disable=SC2016
			GET_MAC='{if ($0 ~ /ether /) { mac = $2 }}'
		elif [ -r /sys/class/net/"$iface"/address ]; then
			MAC=$(cat /sys/class/net/"$iface"/address)
		else
			# shellcheck disable=SC2016
			GET_MAC='{if ($0 ~ /ether /) { mac = $2; } else if ( NR == 1 ) { mac = $5; }}'
		fi
		if [ "$DUPLEX" != 'error' ] && [ "$SPEED" != 'error' ]; then
			$CMD "$iface" | tee -a "$TEE_DEST" | awk "$BEGIN $GET_MAC $GET_ALL $FILL_BLANKS $PRINTF" name="$iface" speed="$SPEED" duplex="$DUPLEX" mac="$MAC"
	 		echo "Cmd = [$CMD $iface];     | awk '$BEGIN $GET_MAC $GET_ALL $FILL_BLANKS $PRINTF' name=$iface speed=$SPEED duplex=$DUPLEX mac=$MAC" >> "$TEE_DEST"
		else
			echo "ERROR: cat command failed for interface $iface" >> "$TEE_DEST"
		fi
	done

elif [ "$KERNEL" = "SunOS" ] ; then
	assertHaveCommandGivenPath /usr/sbin/ifconfig
	assertHaveCommand kstat
	# shellcheck disable=SC2089
	CMD_LIST_INTERFACES="eval /usr/sbin/ifconfig -au | tee $TEE_DEST | egrep -v 'LOOPBACK|netmask' | tee -a $TEE_DEST | grep flags | cut -d':' -f1 | tee -a $TEE_DEST | sort -u | tee -a $TEE_DEST"
	# shellcheck disable=SC2016
	GET_COLLISIONS_RXbytes_TXbytes_SPEED_DUPLEX='($1=="collisions") {collisions=$2} ($1=="duplex" || $1=="link_duplex") {duplex=$2} ($1=="rbytes") {RXbytes=$2} ($1=="obytes") {TXbytes=$2} ($1=="ierrors") {RXerrors=$2} ($1=="oerrors") {TXerrors=$2} ($1=="ifspeed") {speed=$2; speed/=1000000; speed=speed "Mb/s"}'
	# shellcheck disable=SC2016
	GET_IP='/ netmask / {for (i=1; i<=NF; i++) {if ($i == "inet") IPv4 = $(i+1); if ($i == "inet6") IPv6 = $(i+1)}}'
	# shellcheck disable=SC2016
    GET_MAC='{if ($1 == "ether") {split($2, submac, ":"); mac=sprintf("%02s:%02s:%02s:%02s:%02s:%02s", submac[1], submac[2], submac[3], submac[4], submac[5], submac[6])}}'
	FILL_BLANKS='{length(speed) || speed = "<n/a>"; length(duplex) || duplex = "<n/a>"; IPv4 = IPv4 ? IPv4 : "<n/a>"; IPv6 = IPv6 ? IPv6 : "<n/a>"}'
	GET_ALL="$GET_COLLISIONS_RXbytes_TXbytes_SPEED_DUPLEX $GET_IP $GET_MAC $FILL_BLANKS"
	# shellcheck disable=SC2090
	out=$($CMD_LIST_INTERFACES)
	lines=$(echo "$out" | wc -l)
	if [ "$lines" -gt 0 ]; then
		echo "$HEADER"
	fi
	for iface in $out
	do
		echo "Cmd = [$CMD_LIST_INTERFACES]" >> "$TEE_DEST"
		NODE=$(uname -n)
		# shellcheck disable=SC2050
		if [ SOLARIS_8 = false ] && [ SOLARIS_9 = false ] ; then
			CMD_DESCRIBE_INTERFACE="eval kstat -c net -n $iface ; /usr/sbin/ifconfig $iface 2>/dev/null"
		else
			CMD_DESCRIBE_INTERFACE="eval kstat -n $iface ; /usr/sbin/ifconfig $iface 2>/dev/null"
		fi
		$CMD_DESCRIBE_INTERFACE | tee -a "$TEE_DEST" | $AWK "$GET_ALL $FORMAT $PRINTF" name="$iface" node="$NODE"
		echo "Cmd = [$CMD_DESCRIBE_INTERFACE];     | $AWK '$GET_ALL $FORMAT $PRINTF' name=$iface node=$NODE" >> "$TEE_DEST"
	done
elif [ "$KERNEL" = "AIX" ] ; then
	assertHaveCommandGivenPath /usr/sbin/ifconfig
	assertHaveCommandGivenPath /usr/bin/netstat
	# shellcheck disable=SC2089
	CMD_LIST_INTERFACES="eval /usr/sbin/ifconfig -au | tee $TEE_DEST | egrep -v 'LOOPBACK|netmask|inet6|tcp_sendspace' | tee -a $TEE_DEST | grep flags | cut -d':' -f1 | tee -a $TEE_DEST | sort -u | tee -a $TEE_DEST"
	# shellcheck disable=SC2016
	GET_COLLISIONS_RXbytes_TXbytes_SPEED_DUPLEX_ERRORS='($1=="Single"){collisions_s=$4} ($1=="Multiple"){collisions=collisions_s+$4} ($1=="Bytes:") {RXbytes=$4 ; TXbytes=$2} ($1=="Media" && $3=="Running:") {speed=$4"Mb/s" ; duplex=$6} ($1="Transmit" && $2="Errors:") {TXerrors=$3 ; RXerrors=$6}'
	# shellcheck disable=SC2016
	GET_IP='/ netmask / {for (i=1; i<=NF; i++) {if ($i == "inet") IPv4 = $(i+1); if ($i == "inet6") IPv6 = $(i+1)}}'
	# shellcheck disable=SC2016
	GET_MAC='/^Hardware Address:/{mac=$3}'
	FILL_BLANKS='{length(speed) || speed = "<n/a>"; length(duplex) || duplex = "<n/a>"; IPv4 = IPv4 ? IPv4 : "<n/a>"; IPv6 = IPv6 ? IPv6 : "<n/a>"}'
	GET_ALL="$GET_COLLISIONS_RXbytes_TXbytes_SPEED_DUPLEX_ERRORS $GET_IP $GET_MAC $FILL_BLANKS"
	# shellcheck disable=SC2090
	out=$($CMD_LIST_INTERFACES)
	lines=$(echo "$out" | wc -l)
	if [ "$lines" -gt 0 ]; then
		echo "$HEADER"
	fi
	for iface in $out
	do
		echo "Cmd = [$CMD_LIST_INTERFACES]" >> "$TEE_DEST"
		NODE=$(uname -n)
		CMD_DESCRIBE_INTERFACE="eval netstat -v $iface ; /usr/sbin/ifconfig $iface"
		$CMD_DESCRIBE_INTERFACE | tee -a "$TEE_DEST" | $AWK "$GET_ALL $FORMAT $PRINTF" name="$iface" node="$NODE"
		echo "Cmd = [$CMD_DESCRIBE_INTERFACE];     | $AWK '$GET_ALL $FORMAT $PRINTF' name=$iface node=$NODE" >> "$TEE_DEST"
	done
elif [ "$KERNEL" = "Darwin" ] ; then
	assertHaveCommand ifconfig
	assertHaveCommand netstat

	CMD_LIST_INTERFACES='ifconfig -u'
	# shellcheck disable=SC2016
	CHOOSE_ACTIVE='/^[a-z0-9]+: / {sub(":", "", $1); iface=$1} /status: active/ {print iface}'
	# shellcheck disable=SC2016
	UNIQUE='sort -u'
	# shellcheck disable=SC2016
	GET_MAC='{$1 == "ether" && mac = $2}'
	# shellcheck disable=SC2016
	GET_IPv4='{$1 == "inet" && IPv4 = $2}'
	# shellcheck disable=SC2016
	GET_IPv6='{if ($1 == "inet6") {sub("%.*$", "", $2);IPv6 = $2}}'
	# shellcheck disable=SC2016
	GET_SPEED_DUPLEX='{if ($1 == "media:") {gsub("[^0-9]", "", $3); speed=$3 "Mb/s"; sub("-duplex.*", "", $4); sub("<", "", $4); duplex=$4}}'
	# shellcheck disable=SC2016
	GET_RXbytes_TXbytes_COLLISIONS_ERRORS='{
        if ($0 ~ /Name/)
        {
            for (i=1; i<=NF; i++)
            {
                if ($i == "Address") {address_column = i;}
                else if ($i == "Ibytes") {ibytes_column = i;}
                else if ($i == "Ierrs") {ierrs_column = i;}
                else if ($i == "Obytes") {obytes_column = i;}
                else if ($i == "Oerrs") {oerrs_column = i;}
                else if ($i == "Coll") {coll_column = i;}
            }
            flag = 1;
        }

        if(flag == 1){
            if ($address_column == mac)
            {
				(ibytes_column == "") ? RXbytes = "<n/a>"  : RXbytes = $(ibytes_column);
				(ierrs_column == "") ? RXerrors = "<n/a>"  : RXerrors = $(ierrs_column);
				(obytes_column == "") ? TXbytes = "<n/a>"  : TXbytes = $(obytes_column);
				(oerrs_column == "") ? TXerrors = "<n/a>"  : TXerrors = $(oerrs_column);
				(coll_column == "") ? collisions = "<n/a>"  : collisions = $(coll_column);
            }
        }
    }'
	FILL_BLANKS='{length(speed) || speed = "<n/a>"; length(duplex) || duplex = "<n/a>"; IPv4 = IPv4 ? IPv4 : "<n/a>"; IPv6 = IPv6 ? IPv6 : "<n/a>"}'
	GET_ALL="$GET_MAC $GET_IPv4 $GET_IPv6 $GET_SPEED_DUPLEX $GET_RXbytes_TXbytes_COLLISIONS_ERRORS $FILL_BLANKS"
	out=$($CMD_LIST_INTERFACES | tee "$TEE_DEST" | awk "$CHOOSE_ACTIVE" | $UNIQUE | tee -a "$TEE_DEST")
	lines=$(echo "$out" | wc -l)
	if [ "$lines" -gt 0 ]; then
		echo "$HEADER"
	fi
	for iface in $out
	do
		echo "Cmd = [$CMD_LIST_INTERFACES];  | awk '$CHOOSE_ACTIVE' | $UNIQUE" >> "$TEE_DEST"
		CMD_DESCRIBE_INTERFACE="eval ifconfig $iface ; netstat -b -I $iface"
		$CMD_DESCRIBE_INTERFACE | tee -a "$TEE_DEST" | awk "$GET_ALL $PRINTF" name="$iface"
		echo "Cmd = [$CMD_DESCRIBE_INTERFACE];     | awk '$GET_ALL $PRINTF' name=$iface" >> "$TEE_DEST"
	done
elif [ "$KERNEL" = "HP-UX" ] ; then
    assertHaveCommand ifconfig
    assertHaveCommand lanadmin
    assertHaveCommand lanscan
    assertHaveCommand netstat

    CMD='lanscan'
	# shellcheck disable=SC2016
    LANSCAN_AWK='/^Hardware/ {next} /^Path/ {next} {mac=$2; ifnum=$3; ifstate=$4; name=$5; type=$8}'
	# shellcheck disable=SC2016
    GET_IP4='{c="netstat -niwf inet | grep "name; c | getline; close(c); if (NF==10) {next} mtu=$2; IPv4=$4; RXbytes=$5; RXerrors=$6; TXbytes=$7; TXerrors=$8; collisions=$9}'
	# shellcheck disable=SC2016
    GET_IP6='{c="netstat -niwf inet6 | grep "name" "; c| getline; close(c); IPv6=$3}'
	# shellcheck disable=SC2016
    GET_SPEED_DUPLEX='{c="lanadmin -x "ifnum ; c | getline; close(c); if (NF==4) speed=$3"Mb/s"; sub("\-.*", "", $4); duplex=tolower($4)}'
    PRINTF='{printf "%-10s %-17s  %-15s  %-42s %-10s  %-16s %-16s %-16s %-16s %-12s %-12s\n", name, mac, IPv4, IPv6, collisions, RXbytes, RXerrors, TXbytes, TXerrors, speed, duplex}'
    FILL_BLANKS='{length(speed) || speed = "<n/a>"; length(duplex) || duplex = "<n/a>"; IPv4 = IPv4 ? IPv4 : "<n/a>"; IPv6 = IPv6 ? IPv6 : "<n/a>"}'
	out=$($CMD | awk "$LANSCAN_AWK $GET_IP4 $GET_IP6 $GET_SPEED_DUPLEX $PRINTF $FILL_BLANKS")
	lines=$(echo "$out" | wc -l)
	if [ "$lines" -gt 0 ]; then
		echo "$HEADER"
		echo "$out"
	fi
elif [ "$KERNEL" = "FreeBSD" ] ; then
	assertHaveCommand ifconfig
	assertHaveCommand netstat

	CMD_LIST_INTERFACES='ifconfig -a'
	# shellcheck disable=SC2016
	CHOOSE_ACTIVE='/LOOPBACK/ {next} !/RUNNING/ {next} /^[a-z0-9]+: / {sub(":$", "", $1); print $1}'
	UNIQUE='sort -u'
	# shellcheck disable=SC2016
	GET_MAC='{$1 == "ether" && mac = $2}'
	# shellcheck disable=SC2016
	GET_IP='/ netmask / {for (i=1; i<=NF; i++) {if ($i == "inet") IPv4 = $(i+1); if ($i == "inet6") IPv6 = $(i+1)}}'
	# shellcheck disable=SC2016
	GET_SPEED_DUPLEX='/media: / {sub("\134(", "", $4); speed=$4; sub("-duplex.*", "", $5); sub("<", "", $5); duplex=$5}'
	# shellcheck disable=SC2016
	GET_RXbytes_TXbytes_COLLISIONS_ERRORS='{
        if ($0 ~ /Name/)
        {
            for (i=1; i<=NF; i++)
            {
                if ($i == "Address") {address_column = i;}
                else if ($i == "Ibytes") {ibytes_column = i;}
                else if ($i == "Ierrs") {ierrs_column = i;}
                else if ($i == "Obytes") {obytes_column = i;}
                else if ($i == "Oerrs") {oerrs_column = i;}
                else if ($i == "Coll") {coll_column = i;}
            }
            flag = 1;
        }

        if(flag == 1){
            if ($address_column == mac)
            {
				(ibytes_column == "") ? RXbytes = "<n/a>"  : RXbytes = $(ibytes_column);
				(ierrs_column == "") ? RXerrors = "<n/a>"  : RXerrors = $(ierrs_column);
				(obytes_column == "") ? TXbytes = "<n/a>"  : TXbytes = $(obytes_column);
				(oerrs_column == "") ? TXerrors = "<n/a>"  : TXerrors = $(oerrs_column);
				(coll_column == "") ? collisions = "<n/a>"  : collisions = $(coll_column);
            }
        }
    }'
	FILL_BLANKS='{length(speed) || speed = "<n/a>"; length(duplex) || duplex = "<n/a>"; IPv4 = IPv4 ? IPv4 : "<n/a>"; IPv6 = IPv6 ? IPv6 : "<n/a>"}'
	GET_ALL="$GET_MAC $GET_IP $GET_SPEED_DUPLEX $GET_RXbytes_TXbytes_COLLISIONS_ERRORS $FILL_BLANKS"
	out=$($CMD_LIST_INTERFACES | tee "$TEE_DEST" | awk "$CHOOSE_ACTIVE" | $UNIQUE | tee -a "$TEE_DEST")
	lines=$(echo "$out" | wc -l)
	if [ "$lines" -gt 0 ]; then
		echo "$HEADER"
	fi
	for iface in $out
	do
		echo "Cmd = [$CMD_LIST_INTERFACES];  | awk '$CHOOSE_ACTIVE' | $UNIQUE" >> "$TEE_DEST"
		CMD_DESCRIBE_INTERFACE="eval ifconfig $iface ; netstat -b -I $iface"
		$CMD_DESCRIBE_INTERFACE | tee -a "$TEE_DEST" | awk "$GET_ALL $PRINTF" name="$iface"
		echo "Cmd = [$CMD_DESCRIBE_INTERFACE];     | awk '$GET_ALL $PRINTF' name=$iface" >> "$TEE_DEST"
	done
fi
# jscpd:ignore-end
