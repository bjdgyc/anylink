# web
FROM node:lts-alpine as builder_node
WORKDIR /web
COPY ./web /web
RUN npx browserslist@latest --update-db \
    && npm install \
    && npm run build \
    && ls /web/ui

# server
FROM golang:1.16-alpine as builder_golang
#TODO 本地打包时使用镜像
#ENV GOPROXY=https://goproxy.io
ENV GOOS=linux
WORKDIR /anylink
COPY . /anylink
COPY --from=builder_node /web/ui  /anylink/server/ui

#TODO 本地打包时使用镜像
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache git gcc
RUN cd /anylink/server;go build -o anylink -ldflags "-X main.CommitId=$(git rev-parse HEAD)" \
    && /anylink/server/anylink tool -v

# anylink
FROM alpine
LABEL maintainer="github.com/bjdgyc"

ENV IPV4_CIDR="192.168.10.0/24"

WORKDIR /app
COPY --from=builder_golang /anylink/server/anylink  /app/
COPY docker_entrypoint.sh  /app/

COPY ./server/conf  /app/conf
COPY ./server/files  /app/conf/files


#TODO 本地打包时使用镜像
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache bash iptables \
    && chmod +x /app/docker_entrypoint.sh \
    && ls /app

EXPOSE 443 8800

#CMD ["/app/anylink"]
ENTRYPOINT ["/app/docker_entrypoint.sh"]

