#!/bin/ash

#LED
echo 2 > /sys/class/gpio/export
echo 3 > /sys/class/gpio/export
echo 13 > /sys/class/gpio/export
echo 15 > /sys/class/gpio/export

#link
echo 10 > /sys/class/gpio/export

#buttons
echo 0 > /sys/class/gpio/export
echo 1 > /sys/class/gpio/export
echo 16 > /sys/class/gpio/export


while :
do

ButtonReset=`awk '/gpio-0 /{print $5}' /sys/kernel/debug/gpio`
ButtonWPS=`awk '/gpio-1 /{print $5}' /sys/kernel/debug/gpio`
ButtonCheck=`awk '/gpio-16 /{print $5}' /sys/kernel/debug/gpio`
CNCNTLink=`awk '/gpio-10 /{print $5}' /sys/kernel/debug/gpio`

#bia³e
if [ "$ButtonReset" = 'lo' ] && [ "$ButtonWPS" = 'lo' ] && [ "$ButtonCheck" = 'hi' ] ; then
echo high > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
fi

#niebieskie
if [ "$ButtonReset" = 'hi' ] || [ "$ButtonWPS" = 'hi' ] || [ "$ButtonCheck" = 'lo' ] ; then
echo high > /sys/class/gpio/gpio2/direction
echo low > /sys/class/gpio/gpio13/direction
echo low > /sys/class/gpio/gpio15/direction
fi

#zolte
if [ "$ButtonReset" = 'hi' ] && [ "$ButtonWPS" = 'hi' ] ; then
echo low > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
fi
if [ "$ButtonReset" = 'hi' ] && [ "$ButtonCheck" = 'lo' ] ;then
echo low > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
fi
if [ "$ButtonWPS" = 'hi' ] && [ "$ButtonCheck" = 'lo' ] ; then
echo low > /sys/class/gpio/gpio2/direction
echo high > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
fi

#fw side switch
if [ "$ButtonReset" = 'hi' ] && [ "$ButtonWPS" = 'hi' ] && [ "$ButtonCheck" = 'lo' ] ; then
echo low > /sys/class/gpio/gpio2/direction
echo low > /sys/class/gpio/gpio13/direction
echo high > /sys/class/gpio/gpio15/direction
fwupdate sw > /dev/null 2>&1
sync
reboot
fi

if [ "$CNCNTLink" = 'hi' ] ; then
echo low > /sys/class/gpio/gpio3/direction
fi
if [ "$CNCNTLink" = 'lo' ] ; then
echo high > /sys/class/gpio/gpio3/direction
fi

done

exit 0
