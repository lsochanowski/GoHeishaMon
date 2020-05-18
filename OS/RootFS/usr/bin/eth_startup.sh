#!/bin/ash

echo "eth_startup.sh"

lan_ifname=`uci show network.lan.ifname | cut -d "=" -f 2`

if [ "$lan_ifname" = eth0 ] ; then
	uci set network.lan.proto=dhcp
	uci set network.lan.type=
	uci set network.lan.ipaddr=
	uci set network.lan.netmask=
	uci set network.lan.gateway=
	uci set network.lan.dns=
fi

uci set network.lan.ifname=eth1
uci set network.wan.ifname=eth0


exit 1
