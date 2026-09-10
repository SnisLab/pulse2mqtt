#!/bin/bash

if  [ "$1" = install ] 
then
    echo "Installmode is $1"
    if [ -e "/etc/systemd/system/pulse2mqtt.service" ]
    then
        systemctl stop pulse2mqtt.service
        systemctl disable pulse2mqtt.service
        rm /etc/systemd/system/pulse2mqtt.service
    fi
fi
if  [ "$1" = upgrade ] 
then
    echo "Installmode is $1"
    if [ -e "/etc/systemd/system/pulse2mqtt.service" ]
    then
        systemctl stop pulse2mqtt.service
        systemctl disable pulse2mqtt.service
        rm /etc/systemd/system/pulse2mqtt.service
    fi
fi