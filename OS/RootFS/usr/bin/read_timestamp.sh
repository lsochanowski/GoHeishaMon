#!/bin/ash

read_file="/root/powercal"

if [ ! -e "$read_file" ] ; then
	echo "$read_file not found"
	exit 255
fi

timestamp=`sed -n "$1"p $read_file`
timestamp_val=`echo $timestamp | cut -d "=" -f 2`

char_data=`echo $timestamp_val | cut -c $2`

data=`printf "%d" \'$char_data`
#echo "$data"

exit $data
