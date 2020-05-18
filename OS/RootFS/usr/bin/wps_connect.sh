#!/bin/ash

conn_err="*Not connected*"

wifi_link=`iw dev wlan0 link`

prev_mac=`echo "$wifi_link" | grep Connected | cut -d " " -f 3`
prev_ssid=`echo "$wifi_link" | grep SSID | cut -d " " -f 2`

wps_ret=`wpa_cli wps_pbc`
if [ -z "$wps_ret" ] ; then
	echo "Command failed: wpa_cli wps_pbc"
	exit 3
fi
echo "$wps_ret"

wait_time=0

while [ $wait_time -lt 120 ]
do
	sleep 1

	wifi_sts=`iw dev wlan0 link`

	case $wifi_sts in
		$conn_err)
			;;
		*)
			break
			;;
		esac

	wait_time=$(( $wait_time + 1 ))
	echo $wait_time
done


case $wifi_sts in
	$conn_err)
		conn_result=2
		echo "error: Connect failed"
		;;
	*)
		mac=`echo "$wifi_sts" | grep Connected | cut -d " " -f 3`
		ssid=`echo "$wifi_sts" | grep SSID | cut -d " " -f 2`

		if [ "$mac" = "$prev_mac" -a "$ssid" = "$prev_ssid" ] ; then
			conn_result=1
			echo "Previous connection router"
		else
			conn_result=0
		fi
	;;
esac

exit $conn_result
