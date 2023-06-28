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

HEADER='USERNAME                      UID                             GID                             HOME_DIR                                                      USER_INFO'
HEADERIZE="BEGIN {print \"$HEADER\"}"

CMD='cat /etc/passwd'
AWK_IFS='-F:'
# shellcheck disable=SC2016
FILTER='($NF !~ /sh$/) {next}'
# shellcheck disable=SC2016
PRINTF='{printf "%-30.30s %-30.30s %-30.30s  %-60.60s  %s\n", $1, $3, $4, $6, $5}'

if [ "$KERNEL" = "Linux" ] ; then
	# shellcheck disable=SC2016
	FILL_BLANKS='{$5 || $5 = "?"; length($4) || $4 = "?"; length($3) || $3 = "?"}'
elif [ "$KERNEL" = "SunOS" ] ; then
	# shellcheck disable=SC2016
	FILL_BLANKS='{$5 || $5 = "?"; length($4) || $4 = "?"; length($3) || $3 = "?"}'
elif [ "$KERNEL" = "AIX" ] ; then
	# shellcheck disable=SC2016
	FILL_BLANKS='{$5 || $5 = "?"; length($4) || $4 = "?"; length($3) || $3 = "?"}'
elif [ "$KERNEL" = "HP-UX" ] ; then
	# shellcheck disable=SC2016
	FILL_BLANKS='{$5 || $5 = "?"; length($4) || $4 = "?"; length($3) || $3 = "?"}'
elif [ "$KERNEL" = "Darwin" ] ; then
	CMD='dscacheutil -q user'
	AWK_IFS=''
	# shellcheck disable=SC2016
	MASSAGE='/^name: / {username = $2} /^uid: / {UID = $2} /^gid: / {GID = $2} /^dir: / {homeDir = $2} /^shell: / {shell = $2} /^gecos: / {userInfo = $2; for (i=3; i<=NF; i++) userInfo = userInfo " " $i} !/^gecos: / {next}'
	FILTER='{if (shell !~ /sh$/) next; if (homeDir ~ /^[0-9]+$/) next}'
	PRINTF='{printf "%-30.30s %-30.30s %-30.30s %-60.60s  %s\n", username, length(UID) ? UID : "?", length(GID)  ? GID : "?", length(homeDir) ? homeDir : "?", userInfo}'
elif [ "$KERNEL" = "FreeBSD" ] ; then
	# shellcheck disable=SC2016
	FILL_BLANKS='{$5 || $5 = "?"; length($4) || $4 = "?"; length($3) || $3 = "?"}'
fi

assertHaveCommand "$CMD"
# shellcheck disable=SC2086
$CMD | tee "$TEE_DEST" | $AWK $AWK_IFS "$HEADERIZE $MASSAGE $FILTER $FILL_BLANKS $PRINTF"  header="$HEADER"
echo "Cmd = [$CMD];  | $AWK $AWK_IFS '$HEADERIZE $MASSAGE $FILTER $FILL_BLANKS $PRINTF' header=\"$HEADER\"" >> "$TEE_DEST"
