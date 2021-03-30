FROM golang:alpine as builder
ENV GOPROXY=https://goproxy.io \
    GO111MODULE=on \
    GOOS=linux
WORKDIR /root/
RUN apk add --no-cache --update bash git g++ nodejs npm \
    && git clone https://github.com/bjdgyc/anylink.git \
    && cd anylink/server \
    && go build -o anylink -ldflags "-X main.COMMIT_ID=$(git rev-parse HEAD)" \
    && cd ../web \
    && npm install \
    && npx browserslist@latest --update-db \
    && npm run build


FROM golang:alpine
LABEL maintainer="www.mrdoc.fun"
COPY --from=builder /root/anylink/server  /app/
COPY --from=builder /root/anylink/web/ui  /app/ui/
COPY --from=builder /root/anylink/docker /app/
WORKDIR /app
RUN apk add --no-cache pwgen bash iptables openssl ca-certificates \
    && rm -f /app/conf/server.toml \
    && chmod +x docker_entrypoint.sh

ENTRYPOINT ["./docker_entrypoint.sh"]
