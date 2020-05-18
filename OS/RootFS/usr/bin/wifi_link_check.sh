#!/bin/ash

sig=`iw dev wlan0 station dump | grep "signal avg:"`

if [ -z "$sig" ] ; then
	echo "error: WIFI link failed"
	exit 255
fi

rssi=`echo "$sig" | cut -d ":" -f 2 | cut -d "[" -f 1`

echo "RSSI: $rssi"
sig_lvl=$(( $rssi * -1 ))

exit $sig_lvl
