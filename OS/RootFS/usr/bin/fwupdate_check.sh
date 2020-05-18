#!/bin/ash

fw_check=`fwupdate`

cur_side=`echo $fw_check | cut -d ":" -f 6`

echo $cur_side

exit $cur_side
