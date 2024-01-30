#!/bin/bash

#github action release.sh

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

ver=$(cat version)
echo "当前版本 $ver"

mkdir artifact-dist
rm -rf artifact-dist/* app

function archive() {
  os=$1
  arch=$2
  #echo "整理部署文件 $os $arch"

  deploy="anylink-$ver-$os-$arch"
  docker container create --platform $os/$arch --name $deploy bjdgyc/anylink:$ver
  docker cp -a $deploy:/app ./

  rm -rf anylink-deploy
  mv app anylink-deploy
  ls -lh anylink-deploy

  tar zcf ${deploy}.tar.gz anylink-deploy
  mv ${deploy}.tar.gz artifact-dist/
}

echo "copy二进制文件"

archive linux amd64
#archive linux arm64

ls -lh artifact-dist

#注意使用root权限运行
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
