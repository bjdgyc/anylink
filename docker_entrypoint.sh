#!/bin/sh
var1=$1

#set -x

case $var1 in
"bash" | "sh")
  echo $var1
  exec "$@"
  ;;

"tool")
  /app/anylink "$@"
  ;;

*)
  sysctl -w net.ipv4.ip_forward=1
  iptables -t nat -A POSTROUTING -s "${IPV4_CIDR}" -o eth0+ -j MASQUERADE
  iptables -nL -t nat

  exec /app/anylink "$@"
  ;;
esac
