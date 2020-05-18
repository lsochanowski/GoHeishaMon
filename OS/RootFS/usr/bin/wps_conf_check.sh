#!/bin/ash

iw_chk=`iwconfig wlan0 | grep ESSID | cut -d ":" -f 2`

cnf_file="/var/run/wpa_supplicant-wlan0.conf"
file_chk=`cat $cnf_file | grep -c network`

if [ $iw_chk = "off/any" ] ; then
	echo "Check: router connect"
	
fi

######################################################
#network設定が2個以上存在することを確認
#(1個目は現在の接続設定、2個目以降は新たな接続設定)
######################################################
if [ $file_chk -lt 2 ] ; then
	echo "Check: $cnf_file"
	exit 2
fi

#cat $cnf_file

exit 0
