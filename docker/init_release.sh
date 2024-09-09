#!/bin/sh

set -x

#TODO 本地打包时使用镜像
if [[ $CN == "yes" ]]; then
  #sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
  sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
  export GOPROXY=https://goproxy.cn
fi


# docker 启动使用 4.19 以上内核
apk add --no-cache ca-certificates bash iproute2 tzdata iptables

# alpine:3.19 兼容老版 iptables
apk add --no-cache iptables-legacy
rm /sbin/iptables
ln -s /sbin/iptables-legacy /sbin/iptables


chmod +x /app/docker_entrypoint.sh
mkdir /app/log

#备份配置文件
cp -r /app/conf /home/conf-bak

tree /app

uname -a
date -Iseconds
