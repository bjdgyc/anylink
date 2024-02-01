#!/bin/sh

set -x

#TODO 本地打包时使用镜像
if [[ $CN == "yes" ]]; then
  sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
  export GOPROXY=https://goproxy.cn
fi

apk add tzdata gcc musl-dev upx

uname -a
env
date

cd /server

go mod tidy

echo "start build"

extldflags="-static"
ldflags="-s -w -X main.appVer=$appVer -X main.commitId=$commitId -X main.buildDate=$(date -Iseconds) \
  -extldflags \"$extldflags\" "

go build -o anylink -trimpath -ldflags "$ldflags"

ls -lh /server/

# 压缩文件
upx -9 -k anylink

/server/anylink -v
