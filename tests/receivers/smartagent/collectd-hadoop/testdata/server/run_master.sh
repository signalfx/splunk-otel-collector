#!/usr/bin/env bash
set -e

service ssh start

/usr/local/hadoop/bin/hdfs namenode -format
/usr/local/hadoop/sbin/start-dfs.sh
/usr/local/hadoop/sbin/start-yarn.sh
echo "hadoop is running"

sleep infinity
