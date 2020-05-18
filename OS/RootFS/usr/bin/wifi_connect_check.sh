#!/bin/ash

conn_err="*Not connected*"

wifi_link=`iw dev wlan0 link`

case $wifi_link in
	$conn_err)
		#echo "wifi not connected"
		chk_result=1
		;;
	*)
		if [ -z "$wifi_link" ] ; then
			#echo "wifi interface down"
			chk_result=2
		else
			#echo "wifi connected"
			chk_result=0
		fi
		;;
esac

exit $chk_result
