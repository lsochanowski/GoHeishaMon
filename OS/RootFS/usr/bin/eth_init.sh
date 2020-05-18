#!/bin/ash
uci set network.lan.ifname=eth1
uci set network.wan.ifname=eth0

uci set network.lan.type=
uci set network.lan.ipaddr=
uci set network.lan.netmask=
uci set network.lan.gateway=
uci set network.lan.dns=

echo "eth_init.sh"

exit 1
