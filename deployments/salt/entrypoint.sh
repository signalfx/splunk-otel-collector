#!/bin/bash

set -e

echo "export PATH="$PATH:/usr/bin"" >> ~/.bashrc

cat <<-EOF >  /srv/salt/top.sls
base:
  '*':
    - splunk-otel-collector
EOF


cat <<-EOF > /srv/pillar/top.sls
base:
  '*':
    - splunk-otel-collector
EOF

exec "$@"
