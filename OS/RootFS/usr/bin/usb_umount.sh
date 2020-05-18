#!/bin/ash

umount /mnt/usb
rm -rf /mnt/usb

echo high > /sys/class/gpio/gpio11/direction
echo 11 > /sys/class/gpio/unexport

echo "usb_umount.sh"

exit 1
