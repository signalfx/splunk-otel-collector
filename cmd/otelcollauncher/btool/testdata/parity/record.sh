#!/usr/bin/env bash
# Copyright Splunk Inc.
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

# Re-records cases/*/expected from a real `splunk btool` running in the
# splunk/splunk container. Each case directory holds:
#   args      one btool argument per line
#   etc/      files overlaid on a bare Splunk etc directory
#   expected  recorded stdout followed by an EXIT=<code> line
#
# Usage: ./record.sh [case-name...]    (default: all cases)
# Env:   SPLUNK_IMAGE (default splunk/splunk:latest)

set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
image="${SPLUNK_IMAGE:-splunk/splunk:latest}"
name="btool-parity-$$"

if [ "$#" -gt 0 ]; then
    cases=("$@")
else
    cases=()
    for dir in "$here"/cases/*/; do
        cases+=("$(basename "$dir")")
    done
fi

# The entrypoint is replaced so that Splunk itself is never started; btool only
# needs the installed binaries and an etc directory.
docker run -d --rm --name "$name" --platform linux/amd64 --user root \
    --entrypoint tail -e SPLUNK_HOME=/opt/splunk "$image" -f /dev/null >/dev/null
trap 'docker rm -f "$name" >/dev/null 2>&1 || true' EXIT

docker exec "$name" bash -c 'cp -a /opt/splunk-etc/. /opt/splunk/etc/ && chown -R splunk:splunk /opt/splunk/etc /opt/splunk/var'
docker exec "$name" mkdir -p /tmp/parity /tmp/out
docker exec "$name" chown splunk:splunk /tmp/parity /tmp/out
tar -C "$here" -cf - "${cases[@]/#/cases/}" | docker exec -i "$name" tar -xf - -C /tmp/parity

docker exec -i -u splunk "$name" bash -s -- "${cases[@]}" <<'EOF'
set -uo pipefail
# Skeleton etc: everything except stock confs, plus conf.conf which btool requires.
mkdir -p /tmp/skel
(cd /opt/splunk/etc && tar cf - --exclude=./apps --exclude=./deployment-apps --exclude=./disabled-apps \
    --exclude=./system/default --exclude=./system/local .) | tar xf - -C /tmp/skel
mkdir -p /tmp/skel/apps /tmp/skel/system/default /tmp/skel/system/local
cp /opt/splunk/etc/system/default/conf.conf /tmp/skel/system/default/
printf '\nSPLUNK_HOME=/opt/splunk\n' >> /tmp/skel/splunk-launch.conf

for case in "$@"; do
    src="/tmp/parity/cases/$case"
    run="/tmp/run/$case"
    mkdir -p "$(dirname "$run")" "/tmp/out/cases/$case"
    cp -a /tmp/skel "$run"
    [ -d "$src/etc" ] && cp -a "$src/etc/." "$run/"
    mapfile -t args < "$src/args"
    out="$(SPLUNK_ETC="$run" /opt/splunk/bin/splunk btool "${args[@]}" 2>/dev/null)"
    code=$?
    # Paths are normalized the same way the Go test normalizes them, so the
    # expected files do not depend on where the fixtures live.
    {
        [ -n "$out" ] && printf '%s\n' "$out" | sed -E "s#$run#@ETC@#g; s#^(@ETC@[^ ]*) +#\\1 #"
        echo "EXIT=$code"
    } > "/tmp/out/cases/$case/expected"
done
EOF

docker exec "$name" tar -cf - -C /tmp/out cases | tar -xf - -C "$here"
echo "Recorded ${#cases[@]} case(s) from $image"
