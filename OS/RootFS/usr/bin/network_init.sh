#!/bin/ash

echo "network_init.sh"

uci set network.lan.ifname=eth1
uci set network.lan.proto=dhcp
uci set network.lan.type=
uci set network.lan.ipaddr=
uci set network.lan.netmask=
uci set network.lan.gateway=
uci set network.lan.dns=

uci set network.wan.ifname=eth0

uci set network.wlan=interface
uci set network.wlan.proto=dhcp

uci set wireless.radio0.channel=auto
uci set wireless.radio0.disabled=1
uci set wireless.radio0.country=DE

uci set wireless.@wifi-iface[0].network=wlan
uci set wireless.@wifi-iface[0].mode=sta
uci set wireless.@wifi-iface[0].encryption=psk2
uci set wireless.@wifi-iface[0].ssid=OpenWrt
uci set wireless.@wifi-iface[0].key=12345678

uci commit network
uci commit wireless

exit 0
