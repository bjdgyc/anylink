#!/bin/env bash

ver=`cat server/base/app_ver.go | grep APP_VER | awk '{print $3}' | sed 's/"//g'`
echo $ver

#docker login -u bjdgyc

docker build -t bjdgyc/anylink .

docker tag bjdgyc/anylink:latest bjdgyc/anylink:$ver

docker push bjdgyc/anylink:$ver
docker push bjdgyc/anylink:latest

