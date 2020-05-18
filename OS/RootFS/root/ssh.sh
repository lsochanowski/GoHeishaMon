#!/bin/sh
if [ -e /usr/sbin/dropbear ]
then
  echo "jest ssh"
else
 echo "Nie ma ssh instaluje"
/usr/bin/usb_mount.sh
/bin/opkg install /mnt/usb/dropbear_2014.63-2_ar71xx.ipk
/usr/bin/usb_umount.sh
/etc/init.d/dropbear start
/etc/init.d/dropbear enable
/etc/init.d/telnet disable
fi

if [ -e /usr/bin/GoHeishaMon_MIPSUPX ]
then
  echo "jest GoHeisha"
else
echo "Nie ma GoHeisha instaluje"
/usr/bin/usb_mount.sh
cp /mnt/usb/GoHeishaMon_MIPSUPX /usr/bin/GoHeishaMon_MIPSUPX
chmod 755 /usr/bin/GoHeishaMon_MIPSUPX
/usr/bin/usb_umount.sh
fi