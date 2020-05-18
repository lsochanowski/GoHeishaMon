#!/bin/ash

uci set network.lan.proto=dhcp
uci set network.lan.ipaddr=
uci set network.lan.netmask=

uci commit network

/etc/init.d/network restart

exit 0
