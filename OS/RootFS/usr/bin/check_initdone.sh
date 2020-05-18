#!/bin/ash

if [ -e /root/initdone ] ; then
	chk_result=1
else
	chk_result=0
fi

exit $chk_result
