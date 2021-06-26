# 常见问题

### anyconnect 客户端问题
> 客户端请使用群共享文件的版本，其他版本没有测试过，不保证使用正常
> 
> 添加QQ群: 567510628

### OTP 动态码
> 请使用手机安装 freeotp ，然后扫描otp二维码，生成的数字即是动态码

### 远程桌面连接
> 本软件已经支持远程桌面里面连接anyconnect。

### 私有证书问题
> anylink 默认不支持私有证书
> 
> 其他使用私有证书的问题，请自行解决

### dpd timeout 设置问题
```
#客户端失效检测时间(秒) dpd > keepalive
cstp_keepalive = 20
cstp_dpd = 30
mobile_keepalive = 40
mobile_dpd = 50
```
> 以上dpd参数为客户端的超时时间, 如一段时间内，没有数据传输，防火墙会主动关闭连接
> 
> 如经常出现 timeout 的错误信息，应根据当前防火墙的设置，适当减小dpd数值

