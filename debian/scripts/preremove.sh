#!/bin/bash

systemctl daemon-reload

if  [ "$1" = remove ] 
then
    echo "Installmode is $1"
    if [ -e "/etc/systemd/system/pulse2mqtt.service" ]
    then
        systemctl stop pulse2mqtt.service
        systemctl disable pulse2mqtt.service
        rm /etc/systemd/system/pulse2mqtt.service
        systemctl daemon-reload
        systemctl reset-failed
    fi
fi
if  [ "$1" = upgrade ] 
then
    echo "Removemode is $1"
fi
#