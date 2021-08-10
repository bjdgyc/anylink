#!/bin/bash

#################################
# Set up Ethernet bridge on Linux
# Requires: bridge-utils
#################################

#yum install bridge-utils

# Define Bridge Interface
br="anylink0"

# Define physical ethernet interface to be bridged
# with TAP interface(s) above.

eth="eth0"
eth_ip="192.168.10.4/24"
eth_broadcast="192.168.10.255"
eth_gateway="192.168.10.1"


brctl addbr $br
brctl addif $br $eth

ip addr del $eth_ip dev $eth
ip addr add 0.0.0.0 dev $eth
ip link set dev $eth ip promisc on

mac=`cat /sys/class/net/$eth/address`
ip link set up address $mac dev $br
ip addr add $eth_ip broadcast $eth_broadcast dev $br


route add default gateway $eth_gateway








