#!/bin/bash

export MONGO="mongo"
major=$(echo "$MONGO_MAJOR" | cut -d '.' -f1)
if [ $major -gt 5 ]; then
    export MONGO="mongosh"
fi

run_mongo_commandline () {
    nohup $MONGO admin --eval "help" > /dev/null 2>&1

}

sleep 5

nohup mongod --dbpath=/data/db &

run_mongo_commandline

RET=$?

while [[ "$RET" -ne 0 ]]; do
  echo "Waiting for MongoDB to start..."
  run_mongo_commandline
  RET=$?
  sleep 2
done

bash /home/mongodb/scripts/setup_user.sh

mongod --dbpath=/data/db --config mongod.conf
