#node:16-bullseye
#golang:1.20-bullseye
#debian:bullseye-slim
#sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list

# 配合 github action 使用
# 需要先编译出ui文件后 再执行docker编译

# server
FROM golang:1.20-alpine3.19 as builder_golang

ARG CN="no"
ARG appVer="appVer"
ARG commitId="commitId"

ENV TZ=Asia/Shanghai

WORKDIR /server
COPY docker/init_build.sh /tmp/
COPY server/ /server/
COPY web/ui  /server/ui

#RUN apk add gcc musl-dev bash
RUN sh /tmp/init_build.sh


# anylink
FROM alpine:3.19
LABEL maintainer="github.com/bjdgyc"

ARG CN="no"

ENV TZ=Asia/Shanghai
ENV ANYLINK_IN_CONTAINER=true

WORKDIR /app
COPY docker/init_release.sh /tmp/

COPY --from=builder_golang /server/anylink  /app/
COPY docker/docker_entrypoint.sh server/bridge-init.sh ./README.md ./LICENSE /app/
COPY ./deploy /app/deploy
COPY ./index_template  /app/index_template
COPY ./server/conf  /app/conf

#TODO 本地打包时使用镜像
RUN sh /tmp/init_release.sh


EXPOSE 443 8800 443/udp

#CMD ["/app/anylink"]
ENTRYPOINT ["/app/docker_entrypoint.sh"]

