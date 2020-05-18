#!/bin/ash

echo "14" > /sys/class/gpio/export
echo "low" > /sys/class/gpio/gpio14/direction
echo "1" > /sys/class/gpio/gpio14/value

