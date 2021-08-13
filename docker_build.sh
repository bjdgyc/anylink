#!/bin/env bash

ver="0.5.1"

#docker login -u bjdgyc

docker build -t bjdgyc/anylink .

docker tag bjdgyc/anylink:latest bjdgyc/anylink:$ver

docker push bjdgyc/anylink:$ver
docker push bjdgyc/anylink:latest

