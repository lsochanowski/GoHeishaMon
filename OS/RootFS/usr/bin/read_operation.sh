#!/bin/ash

read_file="/root/operation"

data=`sed -n "$1"p "$read_file"`

exit $data
