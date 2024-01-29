#!/bin/bash

ver=$(cat version)
echo $ver

# docker login -u bjdgyc

# docker build -t bjdgyc/anylink -f docker/Dockerfile .

docker build -t bjdgyc/anylink --progress=plain --build-arg CN="yes" --build-arg appVer=$ver \
  --build-arg commitId=$(git rev-parse HEAD) --build-arg buildDate=$(date --iso-8601=seconds) \
  -f docker/Dockerfile .

docker tag bjdgyc/anylink:latest bjdgyc/anylink:$ver
