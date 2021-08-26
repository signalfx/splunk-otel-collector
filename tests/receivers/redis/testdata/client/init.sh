#!/bin/bash

random_str() {
  tr -dc A-Za-z0-9 </dev/urandom | head -c 16
}

db_operations() {
# Checking for the connection, return "PONG" if succeeded

redis-cli -h redis-server ping
while true; do
    sleep .15
    s_one=$(random_str)
    s_two=$(random_str)

    # Setting key value pair in the redis server from the client, return "OK" if succeeded
    redis-cli -h redis-server set "'$s_one'" "'$s_two'"

    # Getting value of the given keys, return value if succeeded
    redis-cli -h redis-server get "'$s_one'"

    # Deleting the key
    redis-cli -h redis-server del "'$s_one'"
done
}
echo "redis client started"
db_operations