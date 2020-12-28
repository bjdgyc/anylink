#!/usr/bin/env bash

git clone https://github.com/bjdgyc/anylink-web.git

cd anylink-web
npm install
npm run build

cd ../
cp -r anylink-web/ui .
go build -o anylink -ldflags "-X main.COMMIT_ID=`git rev-parse HEAD`"

#整理部署文件
mkdir anylink-deploy
cd anylink-deploy

cp -r ../anylink .
cp -r ../conf .
cp -r ../down_files .

#注意使用root权限运行
#sudo ./anylink -conf="conf/server.toml"

