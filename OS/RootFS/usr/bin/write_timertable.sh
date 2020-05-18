#!/bin/ash

write_file="/root/timertable"

if [ $1 -eq 0 ] ; then
	if [ $2 -eq 0 ] ; then
		> $write_file
		echo "reset weeklytimer file"
	fi
fi

zero=00
dow=$1
pattern=$2

set=$3
hour=$4
min=$5
zone1=$6
zone2=$7
tank=$8
mode=$9
zone1_heattemp=$10
zone1_cooltemp=$11
zone2_heattemp=$12
zone2_cooltemp=$13
tank_temp=$14

write_data=`echo $zero$dow$pattern $set $hour $min $zone1 $zone2 $tank $mode $zone1_heattemp $zone1_cooltemp $zone2_heattemp $zone2_cooltemp $tank_temp`

echo $write_data >> $write_file

exit 0
