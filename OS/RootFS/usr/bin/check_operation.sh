#!/bin/ash

read_file="/root/operationstatistics"

if [ ! -e "$read_file" ] ; then
    echo "$read_file not found"
    chk_result=1
else
	chk_result=0
fi

exit $chk_result
