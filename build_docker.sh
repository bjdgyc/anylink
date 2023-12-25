#!/bin/bash

ver=`cat server/base/app_ver.go | grep APP_VER | awk '{print $3}' | sed 's/"//g'`
echo $ver

#docker login -u bjdgyc

#docker build -t bjdgyc/anylink .

docker build -t bjdgyc/anylink --build-arg GitCommitId=$(git rev-parse HEAD) -f docker/Dockerfile .

docker tag bjdgyc/anylink:latest bjdgyc/anylink:$ver

exit 0

docker push bjdgyc/anylink:$ver
docker push bjdgyc/anylink:latest

