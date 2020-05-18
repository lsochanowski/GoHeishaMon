#!/bin/ash

eth_link=`ethtool eth1 | grep "Link detected" | cut -d " " -f 3`

case $eth_link in
	yes)
		conn_result=0
		;;
	*)
		conn_result=1
		;;
esac

echo $conn_result
exit $conn_result
