#!/bin/bash -e

### BEGIN INIT INFO
# Provides: devstack
# Required-Start: $local_fs $network $remote_fs
# Required-Stop: $local_fs $network $remote_fs
# Default-Start:  2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: devstack
# Description: devstack
### END INIT INFO

exec &> >(tr -cd '[:print:]\n' | tee -a /var/log/start-devstack.log)

function wait_for_service() {
    local service=$1
    local timeout=$2
    local start_time=`date +%s`
    echo -n "Waiting for $service to start ..."
    while [ $(expr `date +%s` - $start_time) -lt $timeout ]; do
        if systemctl status $service &>/dev/null; then
            echo " OK"
            return 0
        fi
        sleep 5
    done
    echo " FAILED"
    systemctl status $service
    return 1
}

function wait_for_mysql_proc() {
    local timeout=$1
    local start_time=`date +%s`
    echo -n "Waiting for mysql.proc to be ready ..."
    while [ $(expr `date +%s` - $start_time) -lt $timeout ]; do
        if bash -o pipefail -c 'mysql -uroot -psecret -h127.0.0.1 -e "REPAIR TABLE mysql.proc;" |& grep -q "mysql\.proc.*repair.*status.*OK"'; then
            echo " OK"
            return 0
        fi
        sleep 5
    done
    echo " FAILED"
    mysql -uroot -psecret -h127.0.0.1 -e "REPAIR TABLE mysql.proc;"
    return 1
}

if ! ps -p1 | grep -q systemd; then
    echo "systemd is not PID 1!"
    exit 1
fi

wait_for_service rabbitmq-server 60

wait_for_service mysql 60
mysql -uroot -e "ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY 'secret';" &>/dev/null || true
wait_for_mysql_proc 60

su - stack -c 'cd /opt/stack/devstack && ./stack.sh'
systemctl list-units --no-pager --all devstack@*
