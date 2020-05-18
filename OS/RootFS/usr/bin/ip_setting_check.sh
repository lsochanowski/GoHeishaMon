#!/bin/ash

setting_file="/mnt/usb/settings.txt"

lfcode_del=`cat "$setting_file" | tr -d '\r'`

ipaddr=`echo "$lfcode_del" | grep IP | cut -d "=" -f 2`

if [ ! -z "$ipaddr" ] ; then
	echo "ipaddr OK"
	conn_result=0
else
	echo "error: IP not found"
	conn_result=1
fi

exit $conn_result
