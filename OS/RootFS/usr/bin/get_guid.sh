#!/bin/ash

if [ -e /dev/mtdblock9 ]; then
	guid=`hexdump /dev/mtdblock9 -s 4104 -n 10 | grep 0001008`
else
	echo "error: /dev/mtdblock9 not found"
	exit 1
fi

guid0=`echo $guid | cut -d " " -f 2 | cut -b 1-2`
guid1=`echo $guid | cut -d " " -f 2 | cut -b 3-4`
guid2=`echo $guid | cut -d " " -f 3 | cut -b 1-2`
guid3=`echo $guid | cut -d " " -f 3 | cut -b 3-4`
guid4=`echo $guid | cut -d " " -f 4 | cut -b 1-2`
guid5=`echo $guid | cut -d " " -f 4 | cut -b 3-4`
guid6=`echo $guid | cut -d " " -f 5 | cut -b 1-2`
guid7=`echo $guid | cut -d " " -f 5 | cut -b 3-4`
guid8=`echo $guid | cut -d " " -f 6 | cut -b 1-2`
guid9=`echo $guid | cut -d " " -f 6 | cut -b 3-4`

echo 00 > /root/guid
echo 00 >> /root/guid
echo 07 >> /root/guid
echo $guid0 >> /root/guid
echo $guid1 >> /root/guid
echo $guid2 >> /root/guid
echo $guid3 >> /root/guid
echo $guid4 >> /root/guid
echo $guid5 >> /root/guid
echo $guid6 >> /root/guid
echo $guid7 >> /root/guid
echo $guid8 >> /root/guid
echo $guid9 >> /root/guid

exit 0
