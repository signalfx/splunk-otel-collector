#!/bin/bash -e

exec &> >(tr -cd '[:print:]\n' | tee -a /var/log/stop-devstack.log)

su - stack -c 'cd /opt/stack/devstack && ./unstack.sh'
