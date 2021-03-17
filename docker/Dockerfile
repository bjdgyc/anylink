FROM ubuntu:18.04
WORKDIR /
COPY docker_entrypoint.sh docker_entrypoint.sh
RUN mkdir /anylink && apt update && apt install -y wget iptables tar iproute2
ENTRYPOINT ["/docker_entrypoint.sh"]
#CMD ["/anylink/anylink","-conf=/anylink/conf/server.toml"]
