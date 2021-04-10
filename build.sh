#!/bin/env bash

set -x
function RETVAL() {
  rt=$1
  if [ $rt != 0 ]; then
    echo $rt
    exit 1
  fi
}

#当前目录
cpath=$(pwd)

echo "编译二进制文件"
cd $cpath/server
go build -o anylink -ldflags "-X main.COMMIT_ID=$(git rev-parse HEAD)"
RETVAL $?

echo "编译前端项目"
cd $cpath/web
#国内可替换源加快速度
npm install --registry=https://registry.npm.taobao.org
npm run build --registry=https://registry.npm.taobao.org
#npm install
#npm run build
RETVAL $?

cd $cpath

echo "整理部署文件"
deploy="anylink-deploy"
rm -rf $deploy
mkdir $deploy
mkdir $deploy/log

cp -r server/anylink $deploy
cp -r server/conf $deploy
cp -r server/files $deploy
cp -r server/bridge-init.sh $deploy

cp -r web/ui $deploy

#注意使用root权限运行
#cd anylink-deploy
#sudo ./anylink -conf="conf/server.toml"
