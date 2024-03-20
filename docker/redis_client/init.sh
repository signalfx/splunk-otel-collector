#!/bin/bash

echo "Beginning redis client"

# Checking for the connection, return "PONG" if succeded
redis-cli -h redis-server ping

# Setting key value pair in the redis server from the client, return "OK" if succeded
redis-cli -h redis-server set tempkey tempvalue

echo "redis client started"