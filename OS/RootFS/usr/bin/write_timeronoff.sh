#!/bin/ash

write_file="/root/timeronoff"

onoff=$1

write_data=`echo $onoff`

echo $write_data > $write_file

exit 0
