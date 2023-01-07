#!/bin/ash

write_file="/root/powercal"

if [ $1 -eq 1 ] ; then
	> $write_file
fi

no="$1"
heat_cons="$2"
heat_gene="$3"
cool_cons="$4"
cool_gene="$5"
tank_cons="$6"
tank_gene="$7"
zone1_temp="$8"
zone2_temp="$9"
tank_temp="$10"
outdoor="$11"
multi_od_connection="$12"

echo "$no" >> $write_file
echo "$heat_cons">> $write_file
echo "$heat_gene">> $write_file
echo "$cool_cons">> $write_file
echo "$cool_gene">> $write_file
echo "$tank_cons">> $write_file
echo "$tank_gene">> $write_file
echo "$zone1_temp">> $write_file
echo "$zone2_temp">> $write_file
echo "$tank_temp">> $write_file
echo "$outdoor">> $write_file
echo "$multi_od_connection">> $write_file

exit 0
