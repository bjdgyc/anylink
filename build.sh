#!/bin/bash

#当前目录
cpath=$(pwd)

ver=$(cat version)
echo $ver

#前端编译 仅需要执行一次
#bash ./build_web.sh

bash build_docker.sh

deploy="anylink-deploy-$ver"
docker container rm $deploy
docker container create --name $deploy bjdgyc/anylink:$ver
rm -rf anylink-deploy anylink-deploy.tar.gz
docker cp -a $deploy:/app ./anylink-deploy
tar zcf ${deploy}.tar.gz anylink-deploy


./anylink-deploy/anylink -v


echo "anylink 编译完成，目录: anylink-deploy"
ls -lh anylink-deploy


