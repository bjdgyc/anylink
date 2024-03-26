#!/bin/bash

ver=$(cat version)
echo $ver

# docker login -u bjdgyc

# 生成时间 2024-01-30T21:41:27+08:00
# date -Iseconds

#bash ./build_web.sh

# docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 本地不生成镜像
docker build -t bjdgyc/anylink:latest --no-cache --progress=plain \
  --build-arg CN="yes" --build-arg appVer=$ver --build-arg commitId=$(git rev-parse HEAD) \
  -f docker/Dockerfile .

echo "docker tag latest $ver"
docker tag bjdgyc/anylink:latest bjdgyc/anylink:$ver

exit 0

docker tag bjdgyc/anylink:latest registry.cn-qingdao.aliyuncs.com/bjdgyc/anylink:latest

docker push registry.cn-qingdao.aliyuncs.com/bjdgyc/anylink:latest
