#!/bin/sh

read_file="/root/timeronoff"

if [ ! -e "$read_file" ] ; then
	echo "onoff none"
	exit 255
fi

data=`cat "$read_file"`

exit $data
