#!/bin/bash

#################################
# Set up Ethernet bridge on Linux
# Requires: bridge-utils
#################################

# Define Bridge Interface
br="anylink0"

# Define list of TAP interfaces to be bridged,
# for example tap="tap0 tap1 tap2".
tap="tap0"

# Define physical ethernet interface to be bridged
# with TAP interface(s) above.

eth="eth0"
eth_ip="192.168.10.4"
eth_netmask="255.255.255.0"
eth_broadcast="192.168.10.255"
eth_gateway="192.168.10.1"


brctl addbr $br
brctl addif $br $eth

ifconfig $eth 0.0.0.0 up

mac=`cat /sys/class/net/$eth/address`
ifconfig $br hw ether $mac
ifconfig $br $eth_ip netmask $eth_netmask broadcast $eth_broadcast up

route add default gateway $eth_gateway








