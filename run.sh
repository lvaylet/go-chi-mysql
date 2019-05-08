#!/bin/sh

# Start the Cloud SQL Proxy in the background
./cloud_sql_proxy -instances=${CLOUD_SQL_INSTANCE_CONNECTION_NAME}=tcp:3306 &

# Wait for the Cloud SQL Proxy to spin up
sleep 10

# Start the main REST API server
./main
