#!/bin/bash

ver=$(cat version)
echo $ver

echo "docker tag latest $ver"

docker pull --platform=linux/amd64 bjdgyc/anylink:$ver

docker tag bjdgyc/anylink:$ver registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:latest
docker push registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:latest

docker tag bjdgyc/anylink:$ver registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:$ver
docker push registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:$ver

docker rmi bjdgyc/anylink:$ver

#arm64
docker pull --platform=linux/arm64 bjdgyc/anylink:$ver

docker tag bjdgyc/anylink:$ver registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:arm64v8-latest
docker push registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:arm64v8-latest

docker tag bjdgyc/anylink:$ver registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:arm64v8-$ver
docker push registry.cn-hangzhou.aliyuncs.com/bjdgyc/anylink:arm64v8-$ver

docker rmi bjdgyc/anylink:$ver
