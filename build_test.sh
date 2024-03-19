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
echo $ver

#前端编译 仅需要执行一次
#bash ./build_web.sh

echo "copy二进制文件"

# -tags osusergo,netgo,sqlite_omit_load_extension
flags="-trimpath"
ldflags="-s -w -extldflags '-static' -X main.appVer=$ver -X main.commitId=$(git rev-parse HEAD) -X main.buildDate=$(date --iso-8601=seconds)"
#github action
gopath=/go

dockercmd=$(
  cat <<EOF
sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
apk add gcc g++ musl musl-dev tzdata
export GOPROXY=https://goproxy.cn
go mod tidy
echo "build:"
rm anylink
export CGO_ENABLED=1
go build -v -o anylink $flags -ldflags "$ldflags"
./anylink -v
EOF
)

#使用 musl-dev 编译
docker run -q --rm -v $PWD/server:/app -v $gopath:/go -w /app --platform=linux/amd64 \
  golang:1.20-alpine3.19 sh -c "$dockercmd"

#arm64编译
#docker run -q --rm -v $PWD/server:/app -v $gopath:/go -w /app --platform=linux/arm64 \
#  golang:1.20-alpine3.19 go build -o anylink_arm64 $flags -ldflags "$ldflags"
#exit 0

#cd $cpath

echo "整理部署文件"
rm -rf anylink-deploy anylink-deploy.tar.gz
mkdir anylink-deploy
mkdir anylink-deploy/log

cp -r server/anylink anylink-deploy
cp -r server/conf anylink-deploy

cp -r index_template anylink-deploy
cp -r deploy anylink-deploy
cp -r LICENSE anylink-deploy

tar zcvf anylink-deploy.tar.gz anylink-deploy

#注意使用root权限运行
#cd anylink-deploy
#sudo ./anylink --conf="conf/server.toml"
