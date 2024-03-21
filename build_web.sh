#!/bin/bash

rm -rf web/ui server/ui

docker run -it --rm -v $PWD/web:/app -w /app node:16-alpine \
  sh -c "yarn install --registry=https://registry.npmmirror.com && yarn run build"


cp -r web/ui server/ui
