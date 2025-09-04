#!/bin/bash

message="Distribuidos TP0"

# Estoy tomando el campo que dice SERVER_PORT y luego le sigue un igual y con print $2 me quedo con el numero
SERVER_PORT=$(awk -F'=' '/SERVER_PORT/ {print $2}' server/config.ini | tr -d ' ')

SERVER_IP=$(awk -F'=' '/SERVER_IP/ {print $2}' server/config.ini | tr -d ' ')

response=$(docker run --rm --network tp0_testing_net busybox:latest sh -c "echo '$message' | nc $SERVER_IP $SERVER_PORT")

if [ "$response" = "$message" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi