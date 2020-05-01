This project is to modify Panasonic CZ-TAW1 Firware to send data from heat pump to mqtt  instead of aquarea cloud (there is some POC work proving there is a posiblity to send data concurently to Aquarea Cloud and Mqtt Host using only modified TAW but it's not yet implemented in this project )

## This Project Contains:

- Main software (called GoHeishaMon) responsible for parsing data from heat pump - it's golang implementation of project https://github.com/Egyras/HeishaMon 
every mqtt topics are compatible with this project https://github.com/Egyras/HeishaMon/blob/master/MQTT-Topics.md
and there is two  aditional topics to run command's in system runing the software but it need's another manual.

GoHeishaMon can be used without the CZ-TAW1 module on every platform supported by golang (RaspberyPI,Windows,Linux, OpenWrt routers for example) after connecting it to Heatpump over rs232-ttl interface.
If you need help with this project you can try Slack of Heishamon project there is some people who manage this one :)


- OpenWRT Image with preinstalled GoHeishaMon (and removed A2Wmain due to copyright issues) 

CZ-TAW1 flash memory is divided for two parts called "sides". During update A2Wmain software programing other side then actually it boots ,and change the side just before reboot. In this way, normally in CZ-TAW1 there are two versions of firmware: actual and previous.
Updating firmware with GoHeishaMon we use one side , and we can very easly change the side to A2Wmain (and Smart Cloud) by pressing all three buttons on CZ-TAW1 when GoHeishaMon works ( middle LED will change the color to RED and shortly after this it reboots to orginal SmartCloud.
Unfortunatly from orginal software changing the side without having acces to ssh console is possible only when updating other side was take place succesfully.

Summary: 

It is possible to go back to orginal software (A2Wmain with SmartCluod) very quick , without preparing pendrive ,becouse this solution don't remove firmware with A2Wmain (is still on other  "Side" in the flash).

Even the GoHeishaMon is on other side you can't just change the site in orginal software to GoHeishaMon without acces to console. You have to install GoHeishaMon again. 

For installing GoHeishaMon on CZ-TAW1 you need a clean USB drive FAT32 formatted  (there is a problem with some pendrive vendors if it didin't work try another one) 
copy to usb drive three files :
- openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin
- openwrt-ar71xx-generic-cus531-16M-kernel.bin
- GoHeishaMonConfig.new

GoHeishaMonConfig.new need to be edited according to your needs.

After inserting drive with this files in runing CZ-TAW1 you need to push 3 buttons at once for more tnah 10 seconds until middle LED start changing the colors: green-blue-red. You may also notice the LED blinking on your drive ( if drive have it).

Process of update starts ,and it will take app 3,5min. In the meantime CZ-TAW1 reboots , and after a while you will notice middle LED lights white color , so the GoHeishaMon is running.

In addition  this software enable SSH and web acces on CZ-TAW1 with user: root and password: GoHeishaMonpass ( you should change it!)




---- AUTO BUILD  for MIPS don't work....------
Todo:

- rest of the commands
- queue command from a2wmain 
- flag to point to config file
- manuals 
- tests 

..... more....

