#!/bin/bash

docker run -it --rm -v $PWD/web:/app -w /app node:16-alpine \
  sh -c "yarn install --registry=https://registry.npmmirror.com && yarn run build"

rm -rf server/ui
cp -r web/ui server/ui
