#!/bin/bash


run_mongo_commandline () {
if (( $(echo "$MONGO_MAJOR > 5.9" | bc -l) )); then
    nohup gosu mongodb mongosh admin --eval "help" > /dev/null 2>&1
else
    nohup gosu mongodb mongo admin --eval "help" > /dev/null 2>&1
fi
}

sleep 5

chown -R mongodb:mongodb /home/mongodb

nohup gosu mongodb mongod --dbpath=/data/db &

run_mongo_commandline

RET=$?

while [[ "$RET" -ne 0 ]]; do
  echo "Waiting for MongoDB to start..."
  run_mongo_commandline
  RET=$?
  sleep 2
done

if (( $(echo "${MONGO_MAJOR} > 5.9" | bc -l) )); then
    bash /home/mongodb/scripts/setup_user_newer_ver.sh
else
    bash /home/mongodb/scripts/setup_user.sh
fi

gosu mongodb mongod --dbpath=/data/db --config mongod.conf
