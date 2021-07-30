rm /etc/gh/config
/usr/bin/usb_mount.sh
cp /mnt/usb/GoHeishaMonConfig.new /etc/gh/config
/usr/bin/usb_umount.sh 

reload_config
/etc/init.d/dropbear start

/etc/init.d/uhttpd reload
/etc/init.d/uhttpd enable
/etc/init.d/uhttpd restart
