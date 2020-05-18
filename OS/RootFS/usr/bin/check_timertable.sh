#!/bin/ash

read_file="/root/timertable"

if [ ! -e "$read_file" ] ; then
	exit 1
fi

exit 0
