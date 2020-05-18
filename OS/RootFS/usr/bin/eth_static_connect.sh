#!/bin/ash

setting_file="/mnt/usb/settings.txt"

ipaddr=`cat $setting_file | grep IP | cut -d "=" -f 2`
netmask=`cat $setting_file | grep SUB | cut -d "=" -f 2`
gateway=`cat $setting_file | grep GW | cut -d "=" -f 2`
prefer_dns=`cat $setting_file | grep P_DNS | cut -d "=" -f 2`
sub_dns=`cat $setting_file | grep S_DNS | cut -d "=" -f 2`

uci set network.lan.proto=static

uci set network.lan.ipaddr=$ipaddr
uci set network.lan.netmask=$netmask
uci set network.lan.gateway=$gateway
uci set network.lan.dns=$prefer_dns
if [ ! -z "$prefer_dns" ] ; then
	uci add_list network.lan.dns=$sub_dns
fi

echo "eth_static_connect.sh"

uci commit network

/etc/init.d/network restart

exit 0
