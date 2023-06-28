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

# jscpd:ignore-start
if [ "$KERNEL" = "Linux" ] ; then
	assertHaveCommand df
	CMD='df -h --output=source,fstype,size,used,avail,pcent,itotal,iused,iavail,ipcent,target'
	# shellcheck disable=SC2016
	BEGIN='BEGIN { OFS = "\t" }'
	# shellcheck disable=SC2016
	FILTER_POST='/(devtmpfs|tmpfs)/ {next}'
	# shellcheck disable=SC2016
	PRINTF='
	{
		if($0 ~ /^Filesystem.*/){
			sub("Mounted on","MountedOn",$0);
		}
		match($0,/^(.*[^ ]) +([^ ]+) +([^ ]+) +([^ ]+) +([^ ]+) +([^ ]+) +([^ ]+) +([^ ]+) +([^ ]+) +([^ ]+%|-) +(.*)$/,a);
		if (length(a) != 0)
		{ printf "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", a[1],a[2],a[3],a[4],a[5],a[6],a[7],a[8],a[9],a[10],a[11];}
	}'

elif [ "$KERNEL" = "SunOS" ] ; then
	assertHaveCommandGivenPath /usr/bin/df
	CMD_1='eval /usr/bin/df -n ; /usr/bin/df -g'
	CMD_2='/usr/bin/df -h'

	# shellcheck disable=SC2016
	BEGIN='BEGIN { OFS = "\t" }'
	#Filters out Inode info from df -g output -> inodes = Value just before "total files" & ifree = Value just before "free files"
	# shellcheck disable=SC2016
	INODE_FILTER='
	/^\// {key=$1}
	{
		for(i=1;i<=NF;i++)
		{
			if($i == "total" && $(i+1) == "files")
			{
				inodes=$(i-1)
			}
			if($i == "free" && $(i+1) == "files")
			{
				ifree=$(i-1)
			}
		}
	}
	{if(NR%5==0) sub("\\(.*\\)?", "", key); print "INODE:" key, inodes, ifree}'

	CMD="${CMD_1} | ${AWK} '${INODE_FILTER}'; ${CMD_2}"
	FILTER_PRE='/libc_psr/ {next}'

	#Maps fsType and inode info from the output of INODE_FILTER
	# shellcheck disable=SC2016
	MAP_FS_TO_TYPE='/INODE:/ {MoInodes[$1] = $2; MoIFree[$1] = $3;} /: / {
		for(i=1;i<=NF;i++){
			if($i ~ /^\/.*/)
				keyCol=i;
			else if($i ~ /[a-zA-Z0-9]/)
				valueCol=i;
		}
		if($keyCol ~ /^\/.*:/)
			fsTypes[substr($keyCol,1,length($keyCol)-1)] = $valueCol;
		else
			fsTypes[$keyCol]=$valueCol;
	}'

	#Append Type and Inode headers to the main header and print respective fields from values stored in MAP_FS_TO_TYPE variables
	# shellcheck disable=SC2016
	PRINTF='
	{
		if($0 ~ /^Filesystem.*/){
			for(i=1;i<=NF;i++){
				if($i=="Mounted" && $(i+1)=="on"){
					mountedCol=i;
					sub("Mounted on","MountedOn",$0);
				}
			}
			$(NF+1)="Type";
			$(NF+1)="INodes";
			$(NF+1)="IUsed";
			$(NF+1)="IFree";
			$(NF+1)="IUsePct";

			print $0;
		}
	}
	{
		for(i=1;i<=NF;i++)
		{
			if($i ~ /^\/\S*/ && i==mountedCol && !(fsTypes[$mountedCol]~/(devfs|ctfs|proc|mntfs|objfs|lofs|fd|tmpfs)/) && !($0 ~ /.*\/proc.*/)){
				$(NF+1)=fsTypes[$mountedCol];
				$(NF+1)=MoInodes["INODE:"$mountedCol];
				$(NF+1)=MoInodes["INODE:"$mountedCol]-MoIFree["INODE:"$mountedCol];
				$(NF+1)=MoIFree["INODE:"$mountedCol];

				if(MoInodes["INODE:"$mountedCol]>0)
				{
					$(NF+1)=int(((MoInodes["INODE:"$mountedCol]-MoIFree["INODE:"$mountedCol])*100)/MoInodes["INODE:"$mountedCol])"%";
				}
				else
				{
					$(NF+1)="0";
				}

				print $0;
			}
		}
	}'

elif [ "$KERNEL" = "AIX" ] ; then
	assertHaveCommandGivenPath /usr/bin/df
	CMD='eval /usr/sysv/bin/df -n ; /usr/bin/df -kP -F %u %f %z %l %n %p %m'

	# Normalize Size, Used and Avail columns
	# shellcheck disable=SC2016
	NORMALIZE='
	function fromKB(KB) {
		MB = KB/1024;
		if (MB<1024) return MB "M";
		GB = MB/1024;
		if (GB<1024) return GB "G";
		TB = GB/1024; return TB "T"
	}
	{
		if($0 ~ /^Filesystem.*/){
			for(i=1;i<=NF;i++){
				if($i=="1024-blocks") {sizeCol=i; sizeFlag=1;}
				if($i=="Used") {usedCol=i; usedFlag=1;}
				if($i=="Available") {availCol=i; availFlag=1;}
			}
		}
		if(!($0 ~ /^Filesystem.*/) && sizeFlag==1)
			$sizeCol=fromKB($sizeCol);
		if(!($0 ~ /^Filesystem.*/) && usedFlag==1)
			$usedCol=fromKB($usedCol);
		if(!($0 ~ /^Filesystem.*/) && availFlag==1)
			$availCol=fromKB($availCol);
	}'

	#Maps fsType
	# shellcheck disable=SC2016
	MAP_FS_TO_TYPE='/: / {
		for(i=1;i<=NF;i++){
			if($i ~ /^\/.*/)
				keyCol=i;
			else if($i ~ /[a-zA-Z0-9]/)
				valueCol=i;
		}
		if($keyCol ~ /^\/.*:/)
			fsTypes[substr($keyCol,1,length($keyCol)-1)] = $valueCol;
		else
			fsTypes[$keyCol]=$valueCol;
	}'

	# shellcheck disable=SC2016
	BEGIN='BEGIN { OFS = "\t" }'
	# Append Type and Inode headers to the main header and print respective fields from values stored in MAP_FS_TO_TYPE variables
	# shellcheck disable=SC2016
	PRINTF='
	{
		if($0 ~ /^Filesystem.*/){
			sub("%Iused","IUsePct",$0);
			for(i=1;i<=NF;i++){
				if($i=="Iused") iusedCol=i;
				if($i=="Ifree") ifreeCol=i;

				if($i=="Mounted" && $(i+1)=="on"){
					mountedCol=i;
					sub("Mounted on","MountedOn",$0);
				}
			}
			$(NF+1)="Type";
			$(NF+1)="INodes";
			print $0;
		}
	}
	{
		for(i=1;i<=NF;i++)
		{
			if($i ~ /^\/\S*/ && i==mountedCol && !(fsTypes[$mountedCol]~/(devfs|ctfs|proc|mntfs|objfs|lofs|fd|tmpfs)/) && !($0 ~ /.*\/proc.*/)){
				$(NF+1)=fsTypes[$mountedCol];
				$(NF+1)=$iusedCol+$ifreeCol;
				print $0;
			}
		}
	}'

elif [ "$KERNEL" = "HP-UX" ] ; then
    assertHaveCommand df
    assertHaveCommand fstyp
    CMD='df -Pk'
	# shellcheck disable=SC2016
    MAP_FS_TO_TYPE='{c="fstyp " $1; c | getline ft; close(c);}'
    # shellcheck disable=SC2016
	HEADER='Filesystem\tType\tSize\tUsed\tAvail\tUsePct\tINodes\tIUsed\tIFree\tIUsePct\tMountedOn'
	# shellcheck disable=SC2016
	HEADERIZE='/^Filesystem/ {print header; next}'
	# shellcheck disable=SC2016
    FORMAT='{size=$2; used=$3; avail=$4; usePct=$5; mountedOn=$6; $2=ft; $3=size; $4=used; $5=avail; $6=usePct; $7=mountedOn}'
	# shellcheck disable=SC2016
    FILTER_POST='($2 ~ /^(tmpfs)$/) {next}'
	# shellcheck disable=SC2016
	PRINTF='{printf "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11}'
elif [ "$KERNEL" = "Darwin" ] ; then
	assertHaveCommand mount
	assertHaveCommand df
	CMD='eval mount -t nocddafs,autofs,devfs,fdesc,nfs; df -h -T nocddafs,autofs,devfs,fdesc,nfs'
	# shellcheck disable=SC2016
	BEGIN='BEGIN { OFS = "\t" }'
	#Maps fsType
	# shellcheck disable=SC2016
	MAP_FS_TO_TYPE='/ on / {
		for(i=1;i<=NF;i++){
			if($i=="on" && $(i+1) ~ /^\/.*/)
			{
				key=$(i+1);
			}
			if($i ~ /^\(/)
				value=substr($i,2,length($i)-2);
		}
		fsTypes[key]=value;
	}'
	# Append Type and Inode headers to the main header and print respective fields from values stored in MAP_FS_TO_TYPE variables
	# shellcheck disable=SC2016
	PRINTF='
	{
		if($0 ~ /^Filesystem.*/){
			sub("%iused","IUsePct",$0);

			for(i=1;i<=NF;i++){
				if($i=="iused") iusedCol=i;
				if($i=="ifree") ifreeCol=i;

				if($i=="Mounted" && $(i+1)=="on"){
					mountedCol=i;
					sub("Mounted on","MountedOn",$0);
				}
			}
			$(NF+1)="Type";
			$(NF+1)="INodes";
			print $0;
		}
	}
	{
		for(i=1;i<=NF;i++)
		{
			if($i ~ /^\/dev\/.*s[0-9]+$/){
				sub("^/dev/", "", $i);
				sub("s[0-9]+$", "", $i);
			}
			if($i ~ /^\/\S*/ && i==mountedCol){
				$(NF+1)=fsTypes[$mountedCol];
				$(NF+1)=$iusedCol+$ifreeCol;
				print $0;
			}
		}
	}'

elif [ "$KERNEL" = "FreeBSD" ] ; then
	assertHaveCommand mount
	assertHaveCommand df
	CMD='eval mount -t nodevfs,nonfs,noswap,nocd9660; df -ih -t nodevfs,nonfs,noswap,nocd9660'
	# shellcheck disable=SC2016
	BEGIN='BEGIN { OFS = "\t" }'
	#Maps fsType
	# shellcheck disable=SC2016
	MAP_FS_TO_TYPE='/ on / {
		for(i=1;i<=NF;i++){
			if($i=="on" && $(i+1) ~ /^\/.*/)
			{
				key=$(i+1);
			}
			if($i ~ /^\(/)
				value=substr($i,2,length($i)-2);
		}
		fsTypes[key]=value;
	}'
	# Append Type and Inode headers to the main header and print respective fields from values stored in MAP_FS_TO_TYPE variables
	# shellcheck disable=SC2016
	PRINTF='
	{
		if($0 ~ /^Filesystem.*/){
			sub("%iused","IUsePct",$0);

			for(i=1;i<=NF;i++){
				if($i=="iused") iusedCol=i;
				if($i=="ifree") ifreeCol=i;

				if($i=="Mounted" && $(i+1)=="on"){
					mountedCol=i;
					sub("Mounted on","MountedOn",$0);
				}
			}
			$(NF+1)="Type";
			$(NF+1)="INodes";
			print $0;
		}
	}
	{
		for(i=1;i<=NF;i++)
		{
			if($i ~ /^\/\S*/ && i==mountedCol){
				$(NF+1)=fsTypes[$mountedCol];
				$(NF+1)=$iusedCol+$ifreeCol;
				print $0;
			}
		}
	}'

fi
# jscpd:ignore-end

$CMD | tee "$TEE_DEST" | $AWK "$BEGIN $HEADERIZE $FILTER_PRE $MAP_FS_TO_TYPE $FORMAT $FILTER_POST $NORMALIZE $PRINTF"  header="$HEADER"
echo "Cmd = [$CMD];  | $AWK '$BEGIN $HEADERIZE $FILTER_PRE $MAP_FS_TO_TYPE $FORMAT $FILTER_POST $NORMALIZE $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
