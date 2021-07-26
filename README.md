# AnyLink

[![Go](https://github.com/bjdgyc/anylink/workflows/Go/badge.svg?branch=master)](https://github.com/bjdgyc/anylink/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bjdgyc/anylink)](https://pkg.go.dev/github.com/bjdgyc/anylink)
[![Go Report Card](https://goreportcard.com/badge/github.com/bjdgyc/anylink)](https://goreportcard.com/report/github.com/bjdgyc/anylink)
[![codecov](https://codecov.io/gh/bjdgyc/anylink/branch/master/graph/badge.svg?token=JTFLIIIBQ0)](https://codecov.io/gh/bjdgyc/anylink)
![GitHub release](https://img.shields.io/github/v/release/bjdgyc/anylink)
![GitHub downloads)](https://img.shields.io/github/downloads/bjdgyc/anylink/total)
![LICENSE](https://img.shields.io/github/license/bjdgyc/anylink)

AnyLink 是一个企业级远程办公sslvpn的软件，可以支持多人同时在线使用。

## Repo

> github: https://github.com/bjdgyc/anylink

> gitee: https://gitee.com/bjdgyc/anylink

## Introduction

AnyLink 基于 [ietf-openconnect](https://tools.ietf.org/html/draft-mavrogiannopoulos-openconnect-02)
协议开发，并且借鉴了 [ocserv](http://ocserv.gitlab.io/www/index.html) 的开发思路，使其可以同时兼容 AnyConnect 客户端。

AnyLink 使用TLS/DTLS进行数据加密，因此需要RSA或ECC证书，可以通过 Let's Encrypt 和 TrustAsia 申请免费的SSL证书。

AnyLink 服务端仅在CentOS 7、Ubuntu 18.04测试通过，如需要安装在其他系统，需要服务端支持tun/tap功能、ip设置命令。

## Screenshot

![online](screenshot/online.jpg)

## Installation

> 没有编程基础的同学建议直接下载release包，从下面的地址下载 anylink-deploy.tar.gz
>
> https://github.com/bjdgyc/anylink/releases

> 升级 go version = 1.16
>
> 需要提前安装好 golang 和 nodejs
>
> 使用客户端前，必须申请安全的https证书，不支持私有证书连接

```shell
git clone https://github.com/bjdgyc/anylink.git

cd anylink
sh build.sh

# 注意使用root权限运行
cd anylink-deploy
sudo ./anylink

# 默认管理后台访问地址
# http://host:8800
# 默认账号 密码
# admin 123456

```

## Feature

- [x] IP分配(实现IP、MAC映射信息的持久化)
- [x] TLS-TCP通道
- [x] DTLS-UDP通道
- [x] 兼容AnyConnect
- [x] 基于tun设备的nat访问模式
- [x] 基于tap设备的桥接访问模式
- [x] 支持 [proxy protocol v1](http://www.haproxy.org/download/2.2/doc/proxy-protocol.txt) 协议
- [x] 用户组支持
- [x] 多用户支持
- [x] TOTP令牌支持
- [x] TOTP令牌开关
- [x] 流量控制
- [x] 后台管理界面
- [x] 访问权限管理

## Config

> 默认配置文件内有详细的注释，根据注释填写配置即可。

```shell
# 生成后台密码
./anylink tool -p 123456

# 生成jwt密钥
./anylink tool -s
```

> 数据库配置示例

| db_type  | db_source                                              |
| -------- | ------------------------------------------------------ |
| sqlite3  | ./conf/anylink.db                                      |
| mysql    | user:password@tcp(127.0.0.1:3306)/anylink?charset=utf8 |
| postgres | user:password@localhost/anylink?sslmode=verify-full    |

> 示例配置文件
>
> [conf/server-sample.toml](server/conf/server-sample.toml)

## Setting

> 以下参数必须设置其中之一

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
# 请根据服务器内网网卡替换 eth0
iptables -t nat -A POSTROUTING -s 192.168.10.0/24 -o eth0 -j MASQUERADE
# 如果执行第一个命令不生效，可以继续执行下面的命令
# iptables -A FORWARD -i eth0 -s 192.168.10.0/24 -j ACCEPT
# 查看设置是否生效
iptables -nL -t nat
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

## Systemd

添加 systemd脚本

* anylink 程序目录放入 `/usr/local/anylink-deploy`

systemd 脚本放入：

* centos: `/usr/lib/systemd/system/`
* ubuntu: `/lib/systemd/system/`

操作命令:

* 启动: `systemctl start anylink`
* 停止: `systemctl stop anylink`
* 开机自启: `systemctl enable anylink`

## Docker

1. 获取镜像

   ```bash
   docker pull bjdgyc/anylink:latest
   # 查看帮助命令信息
   docker run -it --rm bjdgyc/anylink -h
   ```

2. 生成密码

   ```bash
   docker run -it --rm bjdgyc/anylink tool -p 123456
   #Passwd:$2a$10$lCWTCcGmQdE/4Kb1wabbLelu4vY/cUwBwN64xIzvXcihFgRzUvH2a
   ```

3. 生成jwt secret

   ```bash
   docker run -it --rm bjdgyc/anylink tool -s
   #Secret:9qXoIhY01jqhWIeIluGliOS4O_rhcXGGGu422uRZ1JjZxIZmh17WwzW36woEbA
   ```

4. 启动容器

   ```bash
   docker run -itd --name anylink --privileged \
       -p 443:443 -p 8800:8800 \
       --restart=always \
       bjdgyc/anylink
   ```

5. 使用自定义参数启动容器

   ```bash
   docker run -itd --name anylink --privileged \
       -p 5443:443 -p 8800:8800 \
       --restart=always
       -v=anylink_volume:/app \
       bjdgyc/anylink \
       -c=/etc/server.toml --issuer="AnyLink SSL VPN" \ 
       -c=/etc/server.toml --default_group="普通用户组" \ # VPN用户组
       -c=/etc/server.toml --files_path="./conf/files" \
       -c=/etc/server.toml --cert_file="./conf/vpn_cert.pem" \
       -c=/etc/server.toml --cert_key="./conf/vpn_cert.key" \
       -c=/etc/server.toml --admin_addr=":8800" \ # WEB管理端口
       -c=/etc/server.toml --log_path="" \ # 留空为标准输出
       -c=/etc/server.toml --log_level=debug \ # 日志级别
       -c=/etc/server.toml --admin_user="rokae" \ # 管理账号明辉岑
       -c=/etc/server.toml --ipv4_cidr=172.17.3.0/24 \ # 客户端网段
       -c=/etc/server.toml --ipv4_gateway=172.17.3.1 \ # DHCP 分配网关
       -c=/etc/server.toml --ipv4_start=172.17.3.10 \  # DHCP 分配IP其实
       -c=/etc/server.toml --ipv4_end=172.17.3.254 \   # DHCP 分配IP结束
       -c=/etc/server.toml --ip_lease=28800 # DHCP租约时长
       # "-c=/etc/server.toml" 更多参数可以参考 -h 命令
   ```

6. 构建镜像

   ```bash
   #获取仓库源码
   git clone https://github.com/bjdgyc/anylink.git
   # 构建镜像
   docker build -t anylink .
   ```

## 常见问题

请前往 [问题地址](question.md) 查看具体信息

## Discussion

![qq.png](screenshot/qq.png)

添加QQ群: 567510628

QQ群共享文件有相关软件下载

## Contribution

欢迎提交 PR、Issues，感谢为AnyLink做出贡献。

注意新建PR，需要提交到dev分支，其他分支暂不会合并。

## Other Screenshot

<details>
<summary>展开查看</summary>


![system.jpg](screenshot/system.jpg)
![setting.jpg](screenshot/setting.jpg)
![users.jpg](screenshot/users.jpg)
![ip_map.jpg](screenshot/ip_map.jpg)
![group.jpg](screenshot/group.jpg)

</details>

## License

本项目采用 MIT 开源授权许可证，完整的授权说明已放置在 LICENSE 文件中。

## Thank

<a href="https://www.jetbrains.com">
    <img src="screenshot/jetbrains.png" width="200" height="200" alt="jetbrains.png" />
</a>
