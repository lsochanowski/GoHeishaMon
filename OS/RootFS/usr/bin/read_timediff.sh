#!/bin/sh

read_file="/root/timediff"

if [ ! -e "$read_file" ] ; then
	echo "timediff none"
	exit 255
fi

data=`cat "$read_file"`

exit $data
