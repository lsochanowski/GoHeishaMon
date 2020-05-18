#!/bin/ash

cnf_file="/var/run/wpa_supplicant-wlan0.conf"
conf=`cat "$cnf_file"`
conf_cnt=`echo "$conf" | wc -l`

network=`echo "$conf" | grep -n network`
network1=`echo "$network" | head -2 | tail -1`
network1_pos=`echo "$network1" | cut -d ":" -f 1`

network2=`echo "$network" | head -3 | tail -1`
if [ -z "$network2" ] ; then
	network2_pos=$conf_cnt
else
	network2_pos=`echo "$network2" | cut -d ":" -f 1`
fi

network3=`echo "$network" | tail -n +4`
if [ -z "$network3" ] ; then
	network3_pos=$conf_cnt
else
	network3_pos=`echo "$network3" | cut -d ":" -f 1`
fi

###############################################
#接続対象であるネットワーク設定の範囲を切り取り
#接続に必要なSSIDとKEYを取得する
###############################################
target_pos=$1

if [ "$target_pos" -gt "$network3_pos" ] ; then
	target=`echo "$conf" | tail +$network3_pos`
elif [ "$target_pos" -gt "$network2_pos" ] ; then
	target=`echo "$conf" | head -$network3_pos | tail +$network2_pos`
elif [ "$target_pos" -gt "$network1_pos" ] ; then
	target=`echo "$conf" | head -$network2_pos | tail +$network1_pos`
fi

echo "target:"
echo "$target"

ssid=`echo "$target" | grep ssid | cut -d "\"" -f 2`
echo "ssid: $ssid"

psk=`echo "$target" | grep psk | cut -d "=" -f 2`
psk_chk=`echo "$psk" | grep "\""`
if [ -z "$psk_chk" ] ; then
	key=$psk
else
	key=`echo $psk | cut -d "\"" -f 2`
fi

echo "key: $key"

uci set wireless.@wifi-iface[0].ssid="$ssid"
uci set wireless.@wifi-iface[0].key=$key

uci set network.wlan=interface
uci set network.wlan.proto=dhcp

uci set wireless.radio0.channel=auto
uci set wireless.radio0.disabled=0
uci set wireless.radio0.country=DE

uci set wireless.@wifi-iface[0].network=wlan
uci set wireless.@wifi-iface[0].mode=sta

pairwise=`echo "$target" | grep pairwise`
pairwise_chk=`echo $pairwise | grep CCMP`

group=`echo "$target" | grep group`
group_chk=`echo $group | grep CCMP`

if [ -z "pairwise_chk" -a "$group_chk" ] ; then
	echo "encryption: TKIP"
	uci set wireless.@wifi-iface[0].encryption=psk2+tkip
else
	echo "encryption: CCMP"
	uci set wireless.@wifi-iface[0].encryption=psk2
fi

uci commit network
uci commit wireless

echo "uci setting OK"

exit 0
