#!/bin/ash

echo "wifi_startup.sh"

uci set network.wlan=interface
uci set network.wlan.proto=dhcp

uci set wireless.radio0.channel=auto
wifi_disable=`uci get wireless.radio0.disabled`
uci set wireless.radio0.disabled=0
uci set wireless.radio0.country=DE

uci set wireless.@wifi-iface[0].network=wlan
uci set wireless.@wifi-iface[0].mode=sta
encryption=`uci get wireless.@wifi-iface[0].encryption`
if [ $encryption != "psk2+tkip" ] ; then
	uci set wireless.@wifi-iface[0].encryption=psk2
fi

key=`uci get wireless.@wifi-iface[0].key`
if [ -z "$key" ] ; then
	uci set wireless.@wifi-iface[0].key=12345678
fi

wifi down

exit $wifi_disable
