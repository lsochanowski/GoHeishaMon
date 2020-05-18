#!/bin/ash

echo "wifi_setting_check.sh"

setting_file="/mnt/usb/settings.txt"

lfcode_del=`cat "$setting_file" | tr -d '\r'`

ssid=`echo "$lfcode_del" | grep SSID | cut -d "=" -f 2`
if [ -z "$ssid" ] ; then
	echo "error: SSID not found"
	exit 1
fi

key=`echo "$lfcode_del" | grep KEY | cut -d "=" -f 2`
if [ -z "$key" ] ; then
	echo "error: KEY not found"
	exit 2
fi

exit 0
