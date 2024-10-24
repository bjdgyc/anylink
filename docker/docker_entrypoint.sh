#!/bin/bash
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
  #sysctl -w net.ipv4.ip_forward=1
  #iptables -t nat -A POSTROUTING -s "${IPV4_CIDR}" -o eth0+ -j MASQUERADE
  #iptables -nL -t nat

  # 启动服务 先判断配置文件是否存在
  if [[ ! -f /app/conf/profile.xml ]]; then
    /bin/cp -r /home/conf-bak/* /app/conf/
    echo "After the configuration file is initialized, the container will be forcibly exited. Restart the container."
    echo "配置文件初始化完成后，容器会强制退出，请重新启动容器。"
    exit 1
  fi

  # 兼容老版本 iptables
  if [[ $IPTABLES_LEGACY == "on" ]]; then
    rm /sbin/iptables
    ln -s /sbin/iptables-legacy /sbin/iptables
  fi

  exec /app/anylink "$@"
  ;;
esac
