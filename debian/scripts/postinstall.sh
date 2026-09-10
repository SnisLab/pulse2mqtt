#!/bin/bash
systemctl daemon-reload

if  [ "$1" = configure ] 
then
    echo "Installmode is $1"
    if [ -e "/etc/systemd/system/pulse2mqtt.service" ]
    then
        systemctl enable pulse2mqtt.service
        systemctl start pulse2mqtt.service
    fi
fi