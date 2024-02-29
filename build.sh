#!/bin/bash

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
yarn install --registry=https://registry.npmmirror.com
yarn run build
RETVAL $?

echo "编译二进制文件"
cd $cpath/server
rm -rf ui
cp -rf $cpath/web/ui .

# -tags osusergo,netgo,sqlite_omit_load_extension
# -tags musl
flags="-v -trimpath"

# -linkmode external
# -extldflags '-static'
ldflags="-s -w -X main.appVer=$ver -X main.commitId=$(git rev-parse HEAD) -X main.date=$(date -Iseconds)"

#国内可替换源加快速度
export GOPROXY=https://goproxy.io
go mod tidy
go build -o anylink $flags -ldflags "$ldflags"

cd $cpath

exit 0

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
