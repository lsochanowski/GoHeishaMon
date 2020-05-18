#!/bin/ash

echo "11" > /sys/class/gpio/export
echo "high" > /sys/class/gpio/gpio11/direction
echo "0" > /sys/class/gpio/gpio11/value

sleep 5

if [ ! -e /mnt/usb ]; then
	mkdir /mnt/usb
	sleep 1
fi

if [ -e /dev/sda1 ]; then
	mount /dev/sda1 /mnt/usb
	echo "usb_mount.sh"
	sleep 1
	if [ -e /mnt/usb/settings.txt ]; then
		mnt_result=0
    else
		echo "error: settings.txt not found"
		mnt_result=2
	fi
else
	echo "error: sda1 not found"
	mnt_result=1
fi

exit $mnt_result
