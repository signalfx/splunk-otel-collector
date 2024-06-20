#!/bin/bash

set -e

echo "************************************************************"
echo "Setting up users..."
echo "************************************************************"

setup_permissions() {
  mongosh <<EOF
  use admin
  db.createUser( { user: "user1",
                 pwd: "user1",
                 customData: { employeeId: 12345 },
                 roles: [ { role: "clusterMonitor", db: "admin" },
                          { role: "readAnyDatabase", db: "admin" },
                          "readWrite"] },
               { w: "majority" , wtimeout: 500000 } );
EOF
}


echo "Configuring ${USER} permissions. . ."
end=$((SECONDS+20))
while [ $SECONDS -lt $end ]; do
    if setup_permissions; then
        echo "Permissions configured!"
        break
    fi
    echo "Trying again in 5 seconds. . ."
    sleep 5
done

# create root user
nohup gosu mongodb mongosh DBNAME --eval "db.createUser({user: 'admin1', pwd: 'pass1', roles:[{ role: 'root', db: 'admin' }, { role: 'read', db: 'local' }]});"

# create app user/database
nohup gosu mongodb mongosh DBNAME --eval "db.createUser({ user: 'NEWUSER', pwd: 'PASSWORD', roles: [{ role: 'readWrite', db: 'admin' }, { role: 'read', db: 'local' }]});"


nohup gosu mongodb mongosh DBNAME --eval "db.createUser({ user: 'testUser', pwd: 'testPass', roles: [{ role: 'clusterMonitor', db: 'admin' }, { role: 'read', db: 'local' }]});"


echo "************************************************************"
echo "Shutting down"
echo "************************************************************"
nohup gosu mongodb mongosh admin --eval "db.shutdownServer();"