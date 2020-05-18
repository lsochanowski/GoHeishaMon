#!/bin/bash

for a in `ls data`; do 
#echo $a
#cat data/$a



while read p; do

pierwszy=`echo $p | awk 'BEGIN { FS="=" } { print $1 }'`
drugi=`echo $p | awk 'BEGIN { FS="=" } { print $2 }'`
echo AllTopics[$a].$pierwszy = $drugi
done <data/$a




done