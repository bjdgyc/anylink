#!/bin/sh

# docker run -it --rm -v $PWD:/app -v /go:/go -w /app --platform=linux/arm64 golang:alpine3.19 sh build_test.sh

set -x

sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
export GOPROXY=https://goproxy.cn


apk add build-base tzdata gcc musl-dev upx

#go build -o anylink
go build -o anylink -ldflags "-s -w -extldflags '-static'"


go env
uname -a

./anylink -v