# AnyLink

AnyLink 是一个企业级远程办公vpn软件，可以支持多人同时在线使用。

## Introduction

AnyLink 基于 [ietf-openconnect](https://tools.ietf.org/html/draft-mavrogiannopoulos-openconnect-02) 协议开发，并且借鉴了 [ocserv](http://ocserv.gitlab.io/www/index.html) 的开发思路，使其可以同时兼容 AnyConnect 客户端。

AnyLink 使用TLS/DTLS进行数据加密，因此需要RSA或ECC证书，可以通过 Let's Encrypt 和 TrustAsia 申请免费的SSL证书。

AnyLink 服务端仅在CentOs7测试通过，如需要安装在其他系统，需要服务端支持tun功能、ip设置命令。

## Installation

```
git clone https://github.com/bjdgyc/anylink.git
cd anylink
go build -o anylink -ldflags "-X main.COMMIT_ID=`git rev-parse HEAD`"
#注意使用root权限运行
sudo ./anylink -conf="conf/server.toml"
```

## Feature

- [x] IP分配
- [x] TLS-TCP通道
- [x] 兼容AnyConnect
- [x] 多用户支持
- [ ] DTLS-UDP通道
- [ ] 后台管理界面
- [ ] 用户组支持
- [ ] TOTP令牌支持
- [ ] 流量控制
- [ ] 访问权限管理

## Config

- [conf/server.toml](https://github.com/bjdgyc/anylink/blob/master/conf/server.toml)
- [conf/user.toml](https://github.com/bjdgyc/anylink/blob/master/conf/user.toml)

## Setting

1. 开启服务器转发
	```
	# flie: /etc/sysctl.conf
	net.ipv4.ip_forward = 1

	#执行如下命令
	sysctl -w net.ipv4.ip_forward=1
	```

2. 设置nat转发规则
	```
	# eth0为服务器内网网卡
	iptables -t nat -A POSTROUTING -s 192.168.10.0/255.255.255.0 -o eth0 -j MASQUERADE
	```
   
3. 使用AnyConnect客户端连接即可


## License

本项目采用 MIT 开源授权许可证，完整的授权说明已放置在 LICENSE 文件中。








