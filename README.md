# AnyLink

[![Go](https://github.com/bjdgyc/anylink/workflows/Go/badge.svg?branch=master)](https://github.com/bjdgyc/anylink/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bjdgyc/anylink)](https://pkg.go.dev/github.com/bjdgyc/anylink)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjdgyc/anylink)](https://goreportcard.com/report/github.com/bjdgyc/anylink)
[![codecov](https://codecov.io/gh/bjdgyc/anylink/branch/master/graph/badge.svg?token=JTFLIIIBQ0)](https://codecov.io/gh/bjdgyc/anylink)

AnyLink 是一个企业级远程办公ssl vpn软件，可以支持多人同时在线使用。

## Repo

> github: https://github.com/bjdgyc/anylink

> gitee: https://gitee.com/bjdgyc/anylink

## Introduction

AnyLink 基于 [ietf-openconnect](https://tools.ietf.org/html/draft-mavrogiannopoulos-openconnect-02)
协议开发，并且借鉴了 [ocserv](http://ocserv.gitlab.io/www/index.html) 的开发思路，使其可以同时兼容 AnyConnect 客户端。

AnyLink 使用TLS/DTLS进行数据加密，因此需要RSA或ECC证书，可以通过 Let's Encrypt 和 TrustAsia 申请免费的SSL证书。

AnyLink 服务端仅在CentOS7、Ubuntu 18.04测试通过，如需要安装在其他系统，需要服务端支持tun/tap功能、ip设置命令。

## Screenshot

![online](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/online.jpg)

## Installation

> 升级 go version = 1.15

```shell
git clone https://github.com/bjdgyc/anylink.git

cd anylink
sh deploy.sh

# 注意使用root权限运行
cd anylink-deploy
sudo ./anylink -conf="conf/server.toml"

# 默认管理后台访问地址
# http://host:8800
```

## Feature

- [x] IP分配(实现IP、MAC映射信息的持久化)
- [x] TLS-TCP通道
- [x] 兼容AnyConnect
- [x] 基于tun设备的nat访问模式
- [x] 基于tap设备的桥接访问模式
- [x] 支持 [proxy protocol v1](http://www.haproxy.org/download/2.2/doc/proxy-protocol.txt) 协议
- [x] 用户组支持
- [x] 多用户支持
- [x] TOTP令牌支持
- [x] 流量控制
- [x] 后台管理界面
- [x] 访问权限管理
  
- [ ] DTLS-UDP通道

## Config

默认配置文件内有详细的注释，根据注释填写配置即可。

```shell
# 生成后台密码
./anylink -passwd 123456

# 生成jwt密钥
./anylink -secret
```

[conf/server.toml](https://github.com/bjdgyc/anylink/blob/master/conf/server.toml)

## Setting

网络模式选择，需要配置 `link_mode` 参数，如 `link_mode="tun"`,`link_mode="tap"` 两种参数。 不同的参数需要对服务器做相应的设置。

建议优先选择tun模式，因客户端传输的是IP层数据，无须进行数据转换。 tap模式是在用户态做的链路层到IP层的数据互相转换，性能会有所下降。 如果需要在虚拟机内开启tap模式，请确认虚拟机的网卡开启混杂模式。

### tun设置

1. 开启服务器转发

 ```shell
 # flie: /etc/sysctl.conf
 net.ipv4.ip_forward = 1

 #执行如下命令
 sysctl -w net.ipv4.ip_forward=1
 ```

2. 设置nat转发规则

```shell
# eth0为服务器内网网卡
iptables -t nat -A POSTROUTING -s 192.168.10.0/255.255.255.0 -o eth0 -j MASQUERADE
```

3. 使用AnyConnect客户端连接即可

### tap设置

1. 创建桥接网卡

```
注意 server.toml 的ip参数，需要与 bridge-init.sh 的配置参数一致
```

2. 修改 bridge-init.sh 内的参数

```
eth="eth0"
eth_ip="192.168.1.4"
eth_netmask="255.255.255.0"
eth_broadcast="192.168.1.255"
eth_gateway="192.168.1.1"
```

3. 执行 bridge-init.sh 文件

```
sh bridge-init.sh
```

## Soft

相关软件下载： QQ群共享文件: 567510628

## Discussion

![qq.png](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/qq.png)

添加QQ群: 567510628

## Contribution

欢迎提交 PR、Issues，感谢为AnyLink做出贡献

## Other Screenshot

<details>
<summary>展开查看</summary>

![system.jpg](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/system.jpg)
![setting.jpg](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/setting.jpg)
![users.jpg](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/users.jpg)
![ip_map.jpg](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/ip_map.jpg)
![group.jpg](https://gitee.com/bjdgyc/anylink/raw/master/screenshot/group.jpg)

</details>

## License

本项目采用 MIT 开源授权许可证，完整的授权说明已放置在 LICENSE 文件中。


## Thank

<img src="screenshot/jetbrains.png" width="200" height="200" alt="jetbrains.png" />





