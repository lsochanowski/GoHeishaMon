#!/bin/ash

TARDIR='/tmp'
GPIO='/sys/class/gpio/gpio11'
KERNEL_BIN='openwrt-ar71xx-generic-cus531-16M-kernel.bin'
USERLAND_BIN='openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin'
TARFILE='download_file.tgz'

ACTION=$1


die()
{
	echo $1

#	if [ -e $GPIO ]; then
#		echo high > /sys/class/gpio/gpio11/direction
#		echo 11 > /sys/class/gpio/unexport
#	fi

	exit 1
}

start_action()
{
#	if [ ! -e $GPIO ]; then
#		echo 11 > /sys/class/gpio/export
#		echo low > /sys/class/gpio/gpio11/direction
#	else
#		die "gpio resource busy"
#	fi
	
	if [ ! -e $TARDIR/$TARFILE ]; then
		die "$TARFILE not found"
	fi

	tar zxvf $TARDIR/$TARFILE -C $TARDIR
	sleep 3s
	KERNEL_IMG_PATH=$(ls $TARDIR/$KERNEL_BIN 2>/dev/null | tail -1)
	USERLAND_IMG_PATH=$(ls $TARDIR/$USERLAND_BIN 2>/dev/null | tail -1)

	if [ "$KERNEL_IMG_PATH" ] && [ "$USERLAND_IMG_PATH" ]; then
		cd $TARDIR
		fwupdate fw-write ./$KERNEL_BIN ./$USERLAND_BIN > /dev/null 2>&1
		if [ $? != 0 ]; then
			die "update failed"
		fi
	else
		die "no image file found"
	fi

#	echo high > /sys/class/gpio/gpio11/direction
#	echo 11 > /sys/class/gpio/unexport

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
