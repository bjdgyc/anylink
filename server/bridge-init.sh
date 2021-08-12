#!/bin/bash

#yum install bridge-utils

# Define Bridge Interface
br="anylink0"

# 请根据sever服务器信息，更新下面的信息
eth="eth0"
eth_ip="192.168.10.4/24"
eth_broadcast="192.168.10.255"
eth_gateway="192.168.10.1"


brctl addbr $br
brctl addif $br $eth

ip addr del $eth_ip dev $eth
ip addr add 0.0.0.0 dev $eth
ip link set dev $eth up promisc on

mac=`cat /sys/class/net/$eth/address`
ip link set dev $br up address $mac promisc on
ip addr add $eth_ip broadcast $eth_broadcast dev $br


route add default gateway $eth_gateway








