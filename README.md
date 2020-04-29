This project is to modify Panasonic CZ-TAW1 Firware to send data from heat pump to mqtt  instead of aquarea cloud (there is some POC work proving there is a posiblity to send data concurently to Aquarea Cloud and Mqtt Host using only modified TAW but it's not yet implemented in this project )

This Project Contain
- Main software (called GoHeishaMon) responsible for parsing data from heat pump - it's golang implementation of project https://github.com/Egyras/HeishaMon 
every mqtt topics are compatible with this project 
and there is two  aditional topics to run command's in system runing the software but it need's another manual 
goheisha can be used without the TAW  on every platform supported by golang (RPI,windows,linux for example) after connecting it to Heatpump over rs232-ttl interface 

- OpenWRT Image with preinstalled GoHeishaMon (and removed a2wmain due to copyright issues) 

this solution didin't remove firmware with a2wmain (there is two "Sides" in the flash)  and in any moment you can bring back orginal firmware by simply pushing all three buttons for around 10 sec's on taw1 but... it will destroy Goheisha Installation and if you want it back you have to reinstall it from scratch

for installing this software you need a clean USB drive  (there is a problem with some pendrive vendors if it didin't work try another one) 
copy to usb drive couple of files :
- openwrt-ar71xx-generic-cus531-16M-rootfs-squashfs.bin
- openwrt-ar71xx-generic-cus531-16M-kernel.bin
- GoHeishaMonConfig.new

GoHeishaMonConfig.new need to be edited according to your needs 

after inserting drive with this files in runing taw you need to push 3 buttons at once for more tnah 10 seconds until middle LED start flashing with 3 colors

in addition  this software enable SSH on TAW1 with password GoHeishaMonpass ( you should change it!)




---- AUTO BUILD  for MIPS don't work....------
Todo:

- rest of the commands
- queue command from a2wmain 
- flag to point to config file
- manuals 
- tests 

..... more....

