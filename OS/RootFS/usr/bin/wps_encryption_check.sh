#!/bin/ash

#echo "wps_encryption_check.sh"

cnf_file="/var/run/wpa_supplicant-wlan0.conf"
conf=`cat "$cnf_file"`
echo "$conf"

network_cnt=`echo "$conf" | grep network | wc -l`

#################################################
#複数鍵対策(Buffalo):接続対象にprotoがないので
#network設定数とproto+鍵なしの合計数とを比較して
#接続設定を判定する
#################################################
mgmt_none_cnt=`echo "$conf" | grep NONE | wc -l`
proto_cnt=`echo "$conf" | grep proto | wc -l`

cmp_num=$(( mgmt_none_cnt + proto_cnt ))
echo "network_cnt:" $network_cnt "cmp_num:" $cmp_num
if [ "$network_cnt" -eq "$cmp_num" ] ; then
	##############
	#通常の場合
	##############
	proto_rsn=`echo "$conf" | grep -n RSN`
	proto_rsn_cnt=`echo "$proto_rsn" | wc -l`
	if [ "$proto_rsn_cnt" -eq 1 ] ; then
		echo "error: Not a WPA2 key"
		exit 0
	fi
	target_pos=`echo "$proto_rsn" | tail -n +2 | cut -d ":" -f 1`
else
	#########################
	#複数鍵(Buffalo)の場合
	#########################
	key_mgmt=`echo "$conf" | grep -n key_mgmt`
	target_pos=`echo "$key_mgmt" | head -2 | tail -1 | cut -d ":" -f 1`
	#echo "key_mgmt:$key_mgmt"
fi
#echo "target_pos:$target_pos"

echo "Check: WPS encryption OK"

exit $target_pos
