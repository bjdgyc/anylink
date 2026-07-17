## 常见问题

### anyconnect 客户端问题

> 客户端请使用群共享文件的版本，其他版本没有测试过，不保证使用正常
>
> 添加QQ群: 567510628

### OTP 动态码

> 请使用手机安装 freeotp ，然后扫描otp二维码，生成的数字即是动态码

### 用户策略问题

> 只要有用户策略，组策略就不生效，相当于覆盖了组策略的配置

### 远程桌面连接

> 本软件已经支持远程桌面里面连接anyconnect。

### 私有证书问题

> anylink 默认不支持私有证书
>
> 其他使用私有证书的问题，请自行解决

### 客户端连接名称

> 客户端连接名称需要修改 [profile.xml](../server/conf/profile.xml) 文件

```xml

<HostEntry>
    <HostName>VPN</HostName>
    <HostAddress>localhost</HostAddress>
</HostEntry>
```

### dpd timeout 设置问题

```yaml
#客户端失效检测时间(秒) dpd > keepalive
cstp_keepalive = 4
cstp_dpd = 9
mobile_keepalive = 7
mobile_dpd = 15
```

> 以上dpd参数为客户端的超时检测时间, 如一段时间内，没有数据传输，防火墙会主动关闭连接
>
> 如经常出现 timeout 的错误信息，应根据当前防火墙的设置，适当减小dpd数值

### 关于审计日志 audit_interval 参数

> 默认值 `audit_interval = 600` 表示相同日志600秒内只记录一次，不同日志首次出现立即记录
>
> 去重key的格式: 16字节源IP地址 + 16字节目的IP地址 + 2字节目的端口 + 1字节协议类型 + 16字节域名MD5

### 反向代理问题

> anylink 仅支持四层反向代理，不支持七层反向代理
>
> 如Nginx请使用 stream模块

```conf
stream {
    upstream anylink_server {
        server 127.0.0.1:8443;
    }
    server {
        listen 443 tcp;
        proxy_timeout 30s;
        proxy_pass anylink_server;
    }
}
```

> nginx实现 共用443端口 示例

```conf
stream {
    map $ssl_preread_server_name $name {
        vpn.xx.com        myvpn;
        default     defaultpage;
    }
    
    # upstream pool
    upstream myvpn {
        server 127.0.0.1:8443;
    }
    upstream defaultpage {
        server 127.0.0.1:8080;
    }
    
    server {
        listen 443 so_keepalive=on;
        ssl_preread on;
        #接收端也需要设置 proxy_protocol
        #proxy_protocol on;
        proxy_pass $name;
    }
}

```

### 性能问题

```
内网环境测试数据
虚拟服务器：  centos7 4C8G
anylink:    tun模式 tcp传输
客户端文件下载速度：240Mb/s
客户端网卡下载速度：270Mb/s
服务端网卡上传速度：280Mb/s
```

> 客户端tls加密协议、隧道header头都会占用一定带宽


### 用户自助修改密码

> AnyLink 内置了用户自助改密页面, 无需额外部署服务。用户浏览器访问 `https://你的VPN地址/pwd` 即可修改自己的密码。

页面需要填写: 用户名、旧密码、新密码。校验规则:

```
1.旧密码只需输入密码(PIN码)本身, 无需附加 OTP 动态码。
2.新密码长度不能少于 6 位, 且不能与旧密码相同。
3.用户名不存在或旧密码错误, 均返回同一提示"用户名或旧密码错误", 防止用户名枚举。
4.改密页面与 VPN 登录共享防爆锁定: 短时间内多次输错旧密码会被锁定, 锁定策略同登录防爆。
```

> 若开启了 `encryption_password`, 新密码会自动加密保存, 与后台设置密码一致。

### 首次登录强制修改密码

> 用于避免管理员设置的初始密码长期不改的安全隐患。开启后, **新建用户**首次连接 VPN 前, 必须先通过自助改密页修改一次密码, 否则登录被拒。

配置方式(`conf/server.toml`):

```yaml
#新建用户首次登录前强制修改密码
force_first_change_pwd = true
#自助改密页地址(内置页填 https://你的VPN地址/pwd)
portal_url = "https://你的VPN地址/pwd"
```

工作流程:

```
1.开启 force_first_change_pwd 后, 通过后台或批量导入新建的用户会被标记为"需改密"。
2.该用户连接 VPN 时, 即使密码正确也会被拒绝, 并提示前往 portal_url 修改密码。
3.用户访问自助改密页 /pwd 修改密码成功后, "需改密"标记自动清除。
4.再次连接 VPN 即可正常登录。
```

> 说明:
>
> 1.仅对开启开关后**新建**的用户生效, 存量老用户不受影响, 可正常登录。
>
> 2.被拒绝登录不会计入防爆失败次数, 不会导致用户被锁定。
>
> 3.管理员在后台编辑用户不会误清"需改密"标记。

### 登录防爆说明

```
1.用户 A 在 IP 1.2.3.4 上尝试登录:
  用户 A 在 IP 1.2.3.4 上尝试登录失败 5 次，触发了该 IP 上的用户 A 锁定 5 分钟。
  在这 5 分钟内，用户 A 从 IP 1.2.3.4 无法进行新的登录尝试。
2.用户 A 更换 IP 到 1.2.3.5 继续尝试登录:
  用户 A 在 IP 1.2.3.5 上继续尝试登录，并且累计失败 20 次，触发了全局用户 A 锁定 5 分钟。
  在这 5 分钟内，用户 A 从任何 IP 地址都无法进行新的登录尝试。
3.IP 1.2.3.4 上多个用户尝试登录:
  如果从 IP 1.2.3.4 上累计有 40 次失败登录尝试（无论来自多少不同的用户），触发了该 IP 的全局锁定 5 分钟。
  在这 5 分钟内，从 IP 1.2.3.4 的所有登录尝试都将被拒绝。

如果在 N 分钟内没有新的失败尝试，失败计数会在 N 分钟后（*_reset_time）重置。
```