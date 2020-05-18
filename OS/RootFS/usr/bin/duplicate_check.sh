#!/bin/ash

setting_file="/mnt/usb/settings.txt"

ipaddr=`cat $setting_file | grep IP | cut -d "=" -f 2`
cur_ip=`ip -f inet -o addr show eth1 | cut -d " " -f 7 | cut -d / -f 1`

duplicate_chk=`arping -D -w 1 -I eth1 "$ipaddr" | grep "Unicast"`

sleep 2

if [ "$cur_ip" = "$ipaddr" ] ; then
	echo "$cur_ip = $ipaddr"
	conn_result=0
elif [ -z "$duplicate_chk" ] ; then
	echo "set ip: $ipaddr"
	conn_result=0
else
	echo "error: Duplicate IP"
	conn_result=1
fi

exit $conn_result
