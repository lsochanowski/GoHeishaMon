#!/bin/ash

ipaddr=`ip -f inet -o addr show wlan0 | cut -d " " -f 7 | cut -d / -f 1`

case $ipaddr in
	*.*.*.*)
		conn_result=0
		echo "IP address OK"
		;;
	*)
		conn_result=1
		echo "IP address failed"
		;;
esac

exit $conn_result
