#!/bin/sh

set -x

#TODO 本地打包时使用镜像
if [[ $CN == "yes" ]]; then
  sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
  export GOPROXY=https://goproxy.cn
fi

apk add --no-cache bash iptables iproute2 tzdata
chmod +x /app/docker_entrypoint.sh
mkdir /app/log

#备份配置文件
cp -r /app/conf /home/conf-bak

tree /app

uname -a
date -Iseconds
