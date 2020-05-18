#!/bin/ash

read_file="/root/powercal"

data=`sed -n "$1"p "$read_file"`

exit $data
