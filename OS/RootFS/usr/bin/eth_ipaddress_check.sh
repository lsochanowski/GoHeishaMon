#!/bin/ash

ipaddr=`ip -f inet -o addr show eth1 | cut -d " " -f 7 | cut -d / -f 1`

case $ipaddr in
	*.*.*.*)
		conn_result=0
		echo "eth connect completed"
		;;
	*)
		conn_result=1
		echo "eth ip get: failed"
		;;
esac

exit $conn_result
