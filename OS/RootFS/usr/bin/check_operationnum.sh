#!/bin/ash

read_file="/root/operation"

data=`sed -n "$1"p "$read_file"`
if [ -z "$data" ] ; then
	echo "$1 count over"
    exit 1
fi
if [ "$data" -ne "$2" ] ; then
	echo "$2 incorrect number"
	exit 2
fi

exit 0
