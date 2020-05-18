#!/bin/ash

echo "wifi_init.sh"

uci set network.wlan=interface
uci set network.wlan.proto=dhcp

uci set wireless.radio0.channel=auto
uci set wireless.radio0.disabled=0
uci set wireless.radio0.country=DE

uci set wireless.@wifi-iface[0].network=wlan
uci set wireless.@wifi-iface[0].mode=sta
uci set wireless.@wifi-iface[0].encryption=psk2
uci set wireless.@wifi-iface[0].ssid=OpenWrt
uci set wireless.@wifi-iface[0].key=12345678

wifi up

exit 1
