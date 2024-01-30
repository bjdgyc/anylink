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
COPY docker/init.sh /tmp/init.sh
COPY server/ /server/
COPY web/ui  /server/ui

#RUN apk add gcc musl-dev bash
RUN sh /tmp/init.sh


# anylink
FROM alpine:3.19
LABEL maintainer="github.com/bjdgyc"

ENV TZ=Asia/Shanghai
ENV ANYLINK_IN_CONTAINER=true

WORKDIR /app
COPY --from=builder_golang /server/anylink  /app/
COPY docker/docker_entrypoint.sh  /app/
COPY ./server/bridge-init.sh /app/
COPY ./server/conf  /app/conf
COPY ./systemd /app/systemd
COPY ./README.md  /app/README.md
COPY ./LICENSE  /app/LICENSE
COPY ./home  /app/home

#TODO 本地打包时使用镜像
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache bash iptables iproute2\
    && chmod +x /app/docker_entrypoint.sh \
    && mkdir /app/log \
    && ls /app

EXPOSE 443 8800

#CMD ["/app/anylink"]
ENTRYPOINT ["/app/docker_entrypoint.sh"]

