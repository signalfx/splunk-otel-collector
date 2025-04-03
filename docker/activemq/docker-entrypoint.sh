#!/bin/bash
set -e

if [ "$1" = "activemq" ]; then
    python /app/entrypoint/Init.py
    chmod 400 ${ACTIVEMQ_CONFIG_DIR}/jmx.password
    exec /usr/bin/supervisord -n -c /etc/supervisor/supervisord.conf
fi

exec $@
