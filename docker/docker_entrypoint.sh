#!/bin/sh
USER="admin"
MM=$(pwgen -1s)
CREATE_USER=1
CONFIG_FILE='/app/conf/server.toml'

if [ $CREATE_USER -eq 1 ]; then
  if [ ! -e $CREATE_USER ]; then
	    MM=$(pwgen -1s)
            touch $CREATE_USER
	    bash /app/generate-certs.sh
            cd /app/conf/ && cp *.crt  /usr/local/share/ca-certificates/
	    update-ca-certificates --fresh
            userpass=$(/app/anylink -passwd "${MM}"| cut -d : -f2)
	    echo "${userpass}"
            jwttoken=$(/app/anylink -secret | cut -d : -f2)
            echo "-- First container startup --user:${USER} pwd:${MM}"
            sed -i "s/admin/${USER}/g" /app/server-example.toml
            sed -i "s/123456/${MM}/g" /app/server-example.toml
            sed -i "s#usertoken#${userpass}#g" /app/server-example.toml
            sed -i "s/jwttoken/${jwttoken}/g" /app/server-example.toml
            else
                        echo "-- Not first container startup --"
  fi

else
                echo "user switch not create"

fi

if [ ! -f $CONFIG_FILE ]; then
echo "#####Generating configuration file#####"
cp /app/server-example.toml /app/conf/server.toml
else
        echo "#####Configuration file already exists#####"
fi

rtaddr=$(grep "cidr" /app/conf/server.toml |awk -F \" '{print $2}')
sysctl -w net.ipv4.ip_forward=1
iptables -t nat -A POSTROUTING -s "${rtaddr}" -o eth0+ -j MASQUERADE
/app/anylink -conf="/app/conf/server.toml"
