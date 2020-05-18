#!/bin/ash

read_file="/root/timertable"

data=`cat "$read_file" | grep "00$1$2" | cut -d " " -f "$3"`

exit $data
