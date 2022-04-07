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

echo "编译前端项目"
cd $cpath/web
#国内可替换源加快速度
#npx browserslist@latest --update-db
#npm install --registry=https://registry.npm.taobao.org
#npm install
#npm run build

yarn install
yarn run build

RETVAL $?

echo "编译二进制文件"
cd $cpath/server
rm -rf ui
cp -rf $cpath/web/ui .
#国内可替换源加快速度
export GOPROXY=https://goproxy.io
go build -v -o anylink -ldflags "-X main.CommitId=$(git rev-parse HEAD)"
RETVAL $?

cd $cpath

echo "整理部署文件"
deploy="anylink-deploy"
rm -rf $deploy ${deploy}.tar.gz
mkdir $deploy

cp -r server/anylink $deploy
cp -r server/bridge-init.sh $deploy
cp -r server/conf $deploy

cp -r systemd $deploy
cp -r LICENSE $deploy

tar zcvf ${deploy}.tar.gz $deploy

#注意使用root权限运行
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
