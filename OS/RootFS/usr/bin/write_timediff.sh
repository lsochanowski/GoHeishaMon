#!/bin/ash

write_file="/root/timediff"

timediff=$1

write_data=`echo $timediff`

echo $write_data > $write_file

exit 0
