#!/bin/ash

echo "eth test start"

while true
do
    sleep 30
    ifconfig eth1 down
done
