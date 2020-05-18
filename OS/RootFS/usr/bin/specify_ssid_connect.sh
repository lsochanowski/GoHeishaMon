#!/bin/ash

echo "specify_ssid_connect.sh"

setting_file="/mnt/usb/settings.txt"

ssid=`cat $setting_file | grep SSID | cut -d "=" -f 2`
key=`cat $setting_file | grep KEY | cut -d "=" -f 2`

uci set wireless.@wifi-iface[0].ssid="$ssid"
uci set wireless.@wifi-iface[0].key=$key

uci commit network
uci commit wireless

wifi up

exit 0
