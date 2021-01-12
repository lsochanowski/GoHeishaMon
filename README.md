# -!!!!!!! Latest checked release is 1.0.166 !!!!! Others are tests ,and some of them can brick CZ-TAW1.!!!!!!!-


This project is to modify Panasonic CZ-TAW1 Firmware to send data from heat pump to MQTT instead to Aquarea Cloud (there is some POC work proving there is a posiblity to send data concurently to Aquarea Cloud and MQTT host using only modified CZ-TAW1 ,but it's not yet implemented in this project )

### This Project Contains:

- Main software (called GoHeishaMon) responsible for parsing data from Heat Pump - it's golang implementation of project https://github.com/Egyras/HeishaMon 
All MQTT topics are compatible with HeishaMon project: https://github.com/Egyras/HeishaMon/blob/master/MQTT-Topics.md
and there are two aditional topics to run command's in system runing the software but it need's another manual.

GoHeishaMon can be used without the CZ-TAW1 module on every platform supported by golang (RaspberyPi, Windows, Linux, OpenWrt routers for example) after connecting it to Heat Pump over rs232-ttl interface.
If you need help with this project you can try Slack of Heishamon project there is some people who manage this one :)

- OpenWRT Image with preinstalled GoHeishaMon (and removed A2Wmain due to copyright issues) 

CZ-TAW1 flash memory is divided for two parts called "sides". During Smart Cloud update A2Wmain software programing other side then actually it boots ,and change the side just before reboot. In this way, normally in CZ-TAW1 there are two versions of firmware: actual and previous.
Updating firmware with GoHeishaMon we use one side , and we can very easly change the side to Smart Cloud (A2Wmain software) by pressing all three buttons on CZ-TAW1 when GoHeishaMon works ( middle LED will change the color to RED and shortly after this it reboots to orginal SmartCloud).
Unfortunatly from Smart Cloud software changing the side without having acces to ssh console is possible only when updating other side was take place succesfully.

Summary: 

It is possible to go back to orginal software (A2Wmain with SmartCloud) very quick , without preparing pendrive ,becouse this solution don't remove firmware with A2Wmain (is still on other "Side" in the flash).

Even the GoHeishaMon is on other side you can't just change the site in orginal software to GoHeishaMon without acces to console. You have to install GoHeishaMon again. 

## Installation

For installing GoHeishaMon on CZ-TAW1 you need a clean USB drive FAT32 formatted  (there is a problem with some pendrive vendors if it didin't work try another one, becouse of big drop of voltage on USB port please use USB flash memory stick.) https://github.com/lsochanowski/GoHeishaMon/releases/tag/1.0.166
copy to usb drive files :
- openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin
- openwrt-ar71xx-generic-cus531-16M-kernel.bin
- GoHeishaMonConfig.new ( It is config.example file edited according to your needs and changed it's name. Please pay attantion on file extension ,since in Windows .txt is often added)


After inserting drive with this files in runing CZ-TAW1 you need to push 3 buttons at once for more tnah 10 seconds until middle LED start changing the colors: green-blue-red. You may also notice the LED blinking on your drive ( if drive have it).

Process of update starts ,and it will take app 3 min. In the meantime CZ-TAW1 reboots , and after a while you will notice middle LED lights white color . Wait with removing drive from module untill the white LED turn off again ( that is a sign , that GoHeishaMon copied config file from drive and reboot CZ-TAW1. You need to remove the drive before the white LED turn on again , becouse the config file will be copied again and reboot if the drive with a config file will be still present.

### SSH and web (over LuCI) access (on by default since 1.1.159 - to be veryfied)

For advanced users there is possibility to have SSH and web acces (LuCI) on CZ-TAW1:
- In config file you should have option "EnableCommand=true"
- GoHeishaMon should be connected to MQTT server
- Public in MQTT topic "panasonic_heat_pump/OSCommand" (or eqvivalent with is set as Mqtt_set_base) one by one values: "umount /overlay" , "jffs2reset -y" and finally "reboot". This will perform a so called firstboot. You can see the output console in topic"panasonic_heat_pump/OSCommand/out". All configuration ( also including WiFi connection , GoHeishaMon config) will be set to default , so please connect GoHeishaMon via Ethernet cable after that, and use a drive ( or ssh connection and edit file /etc/gh/config) to set GoHeishaMon configuration.  WiFi configuration you can do via ssh or LuCI ,identical to standard OpenWRT routers ( It is alsp posibility ,that CZ-TAW1 will be also a reapeter , or dummy AP ).

After reboot you should be able to connect to ssh and via web with user: root and password: GoHeishaMonpass ( you should change it!)


Screenshot from Homeassistant:
![Screenshot from Homeassistant](PompaCieplaScreen.PNG)



Changes:

-
1.1.166 - add new topics

1.1.159 comparing to1.1.150 :
- removed a2wmain watch
- start ssh and www from script
- Home Assistant MQTT Discovery https://www.home-assistant.io/docs/mqtt/discovery/

1.1.150 comparing to 1.1.135 : 
- moved buttons handling from GoHeishaMon to separate script ( in this way , if GoHeishaMon will crash it is still possible to go back to orginal via 3 buttons)


Todo:

- queue command from a2wmain 
- flag to point to config file
- manuals 
- tests 

..... more....
