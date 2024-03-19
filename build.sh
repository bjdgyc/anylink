#!/bin/bash

#当前目录
cpath=$(pwd)

ver=$(cat version)
echo $ver

#前端编译 仅需要执行一次
bash ./build_web.sh

cd $cpath/server

go build -v -o anylink

./anylink -v


echo "anylink 编译完成，目录: $cpath/server/anylink"


