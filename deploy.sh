#!/usr/bin/env bash

#编译二进制文件
go build -o anylink -ldflags "-X main.COMMIT_ID=`git rev-parse HEAD`"

#编译前端项目
git clone https://github.com/bjdgyc/anylink-web.git
cd anylink-web
#国内可替换源加快速度
#npm install --registry=https://registry.npm.taobao.org
#npm run build --registry=https://registry.npm.taobao.org
npm install
npm run build

cd ../

#整理部署文件
mkdir anylink-deploy
mkdir anylink-deploy/log

cp -r anylink anylink-deploy
cp -r anylink-web/ui anylink-deploy
cp -r conf anylink-deploy
cp -r down_files anylink-deploy

#注意使用root权限运行
#cd anylink-deploy
#sudo ./anylink -conf="conf/server.toml"

