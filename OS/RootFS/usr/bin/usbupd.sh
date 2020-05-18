#!/bin/ash

BINDIR='/tmp'
MOUNTDIR='/mnt/usb'
GPIO='/sys/class/gpio/gpio11'
KERNEL_BIN='openwrt-ar71xx-generic-cus531-16M-kernel.bin'
USERLAND_BIN='openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin'

ACTION=$1
ver=$2

guid=`hexdump /dev/mtdblock9 -s 4104 -n 10 | grep 0001008`
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


if [ ${guid0} -lt 10 ]; then
	guid0=`echo ${guid0}|cut -c 2`
else
	guid0=`(printf "\x${guid0}")`
fi

if [ ${guid1} -lt 10 ]; then
	guid1=`echo ${guid1}|cut -c 2`
else
	guid1=`(printf "\x${guid1}")`
fi

if [ ${guid2} -lt 10 ]; then
	guid2=`echo ${guid2}|cut -c 2`
else
	guid2=`(printf "\x${guid2}")`
fi

if [ ${guid3} -lt 10 ]; then
	guid3=`echo ${guid3}|cut -c 2`
else
	guid3=`(printf "\x${guid3}")`
fi

if [ ${guid4} -lt 10 ]; then
	guid4=`echo ${guid4}|cut -c 2`
else
	guid4=`(printf "\x${guid4}")`
fi

if [ ${guid5} -lt 10 ]; then
	guid5=`echo ${guid5}|cut -c 2`
else
	guid5=`(printf "\x${guid5}")`
fi

if [ ${guid6} -lt 10 ]; then
	guid6=`echo ${guid6}|cut -c 2`
else
	guid6=`(printf "\x${guid6}")`
fi

if [ ${guid7} -lt 10 ]; then
	guid7=`echo ${guid7}|cut -c 2`
else
	guid7=`(printf "\x${guid7}")`
fi

if [ ${guid8} -lt 10 ]; then
	guid8=`echo ${guid8}|cut -c 2`
else
	guid8=`(printf "\x${guid8}")`
fi

if [ ${guid9} -lt 10 ]; then
	guid9=`echo ${guid9}|cut -c 2`
else
	guid9=`(printf "\x${guid9}")`
fi

guidsum=007${guid0}${guid1}${guid2}${guid3}${guid4}${guid5}${guid6}${guid7}${guid8}${guid9}





die()
{
	echo $1
    if [ "$(mount | grep /mnt/usb | grep -v grep)" ]; then
        umount $MOUNTDIR
    fi
    if [ -e $MOUNTDIR ]; then
        rmdir $MOUNTDIR
    fi
    if [ -e $GPIO ]; then
        echo high > /sys/class/gpio/gpio11/direction
	    echo 11 > /sys/class/gpio/unexport
    fi
    exit 1
}

start_action()
{
	if [ ! -e $GPIO ]; then
		echo 11 > /sys/class/gpio/export
		echo low > /sys/class/gpio/gpio11/direction
	else
		die "gpio resource busy"
	fi
	
    if [ ! -e $MOUNTDIR ]; then
        mkdir $MOUNTDIR
    fi
	
	sleep 5
    MOUNT_SUCCESS=no
    mount  /dev/sda1 $MOUNTDIR > /dev/null 2>&1
    if [ $? = 0 ]; then
        MOUNT_SUCCESS=yes
    fi
    if [ "$MOUNT_SUCCESS" != "yes" ]; then
        die "failed to mount /dev/$DEVICE to $MOUNTDIR"
    fi

    touch $MOUNTDIR/${ver}_${guidsum}.txt	


    KERNEL_IMG_PATH=$(ls $MOUNTDIR/$KERNEL_BIN 2>/dev/null | tail -1)
    USERLAND_IMG_PATH=$(ls $MOUNTDIR/$USERLAND_BIN 2>/dev/null | tail -1)

    if [ "$KERNEL_IMG_PATH" ] && [ "$USERLAND_IMG_PATH" ]; then
    	cd $BINDIR
    	cp $KERNEL_IMG_PATH .
    	cp $USERLAND_IMG_PATH .
	
    	fwupdate fw-write ./$KERNEL_BIN ./$USERLAND_BIN > /dev/null 2>&1
	
        if [ $? != 0 ]; then
	        die "update failed"
        fi
    else
        die "no image file found"
    fi

    touch $MOUNTDIR/${guidsum}.txt
    umount $MOUNTDIR || die "umount $MOUNTDIR failed"
    rmdir $MOUNTDIR || die "rmdir $MOUNTDIR failed"
    echo high > /sys/class/gpio/gpio11/direction
    echo 11 > /sys/class/gpio/unexport

    exit 0
}

case $ACTION in
    start)
        start_action
        ;;
    *)
    	echo "unknown action"
        ;;
esac
