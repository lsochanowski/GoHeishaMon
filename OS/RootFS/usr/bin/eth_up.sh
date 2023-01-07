#!/bin/ash

ifconfig wlan0 down
ifconfig eth1 up
sleep 2

#ifconfig

/etc/init.d/network reload
