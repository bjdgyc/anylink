#!/bin/bash

ver=$(cat version)
echo $ver

# docker login -u bjdgyc

# 生成时间 2024-01-30T21:41:27+08:00
# date -Iseconds

docker buildx build -t bjdgyc/anylink:latest --progress=plain --build-arg CN="yes" --build-arg appVer=$ver \
  --build-arg commitId=$(git rev-parse HEAD) -f docker/Dockerfile .

echo "docker tag latest $ver"
docker tag bjdgyc/anylink:latest bjdgyc/anylink:$ver
