#!/bin/ash

ifconfig eth1 down
ifconfig wlan0 up
sleep 5

#ifconfig

/etc/init.d/network reload
