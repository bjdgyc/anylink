#node:16-bullseye
#golang:1.20-bullseye
#debian:bullseye-slim
#sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list


# web
FROM node:16-alpine3.18 as builder_node
WORKDIR /web
COPY ./web /web
RUN yarn install \
    && yarn run build \
    && ls /web/ui


# server
FROM golang:1.20-alpine3.18 as builder_golang
#TODO 本地打包时使用镜像
ENV GOPROXY=https://goproxy.cn
ENV GOOS=linux
ARG GitCommitId="gitCommitId"

WORKDIR /anylink
COPY server /anylink
COPY --from=builder_node /web/ui  /anylink/ui

#TODO 本地打包时使用镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add gcc musl-dev
RUN cd /anylink;go mod tidy;go build -o anylink -ldflags "-s -w -X main.CommitId=${GitCommitId}" \
    && /anylink/anylink tool -v


# anylink
FROM alpine:3.18
LABEL maintainer="github.com/bjdgyc"

ENV ANYLINK_IN_CONTAINER=true

WORKDIR /app
COPY --from=builder_golang /anylink/anylink  /app/
COPY docker/docker_entrypoint.sh  /app/
COPY ./server/bridge-init.sh /app/
COPY ./server/conf  /app/conf
COPY ./LICENSE  /app/LICENSE
COPY ./home  /app/home

#TODO 本地打包时使用镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache bash iptables \
    && chmod +x /app/docker_entrypoint.sh \
    && ls /app

EXPOSE 443 8800

#CMD ["/app/anylink"]
ENTRYPOINT ["/app/docker_entrypoint.sh"]

