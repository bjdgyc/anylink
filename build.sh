#!/bin/bash

github_action=$1

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

#ver=`cat server/base/app_ver.go | grep APP_VER | awk '{print $3}' | sed 's/"//g'`
ver=$(cat version)
echo "当前版本 $ver"

echo "编译前端项目"
cd $cpath/web
#国内可替换源加快速度
#npx browserslist@latest --update-db
if [ "$github_action" == "github_action" ]; then
  yarn install --registry=https://registry.npmmirror.com
else
  yarn install
fi

yarn run build

RETVAL $?

echo "编译二进制文件"
cd $cpath/server
rm -rf ui
cp -rf $cpath/web/ui .

flags="-v -trimpath -extldflags '-static' -tags osusergo,netgo,sqlite_omit_load_extension"
ldflags="-s -w -X main.appVer=$ver -X main.commitId=$(git rev-parse HEAD) -X main.date=$(date --iso-8601=seconds)"

if [ "$github_action" == "github_action" ]; then
  echo "github_action"
else
  #国内可替换源加快速度
  export GOPROXY=https://goproxy.io
  go mod tidy
  go build -o anylink "$flags" -ldflags "$ldflags"
  exit 0
fi

#github action
go mod tidy
go build -o anylink_amd64 "$flags" -ldflags "$ldflags"

#arm64交叉编译
CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ go build -o anylink_arm64 "$flags" -ldflags "$ldflags"

./anylink_amd64 -v
./anylink_arm64 -v

exit 0

cd $cpath

echo "整理部署文件"
deploy="anylink-deploy"
rm -rf $deploy ${deploy}.tar.gz
mkdir $deploy
mkdir $deploy/log

cp -r server/anylink $deploy
cp -r server/bridge-init.sh $deploy
cp -r server/conf $deploy

cp -r systemd $deploy
cp -r LICENSE $deploy
cp -r home $deploy

tar zcvf ${deploy}.tar.gz $deploy

#注意使用root权限运行
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
