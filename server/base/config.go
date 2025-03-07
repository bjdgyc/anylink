package base

const (
	cfgStr = iota
	cfgInt
	cfgBool

	defaultJwt = "abcdef.0123456789.abcdef"
	defaultPwd = "$2a$10$UQ7C.EoPifDeJh6d8.31TeSPQU7hM/NOM2nixmBucJpAuXDQNqNke"
)

type config struct {
	Typ     int
	Name    string
	Short   string
	Usage   string
	ValStr  string
	ValInt  int
	ValBool bool
}

var configs = []config{
	{Typ: cfgStr, Name: "conf", Usage: "config file", ValStr: "./conf/server.toml", Short: "c"},
	{Typ: cfgStr, Name: "profile", Usage: "profile.xml file", ValStr: "./conf/profile.xml"},
	{Typ: cfgStr, Name: "profile_name", Usage: "profile name(用于区分不同服务端的配置)", ValStr: "anylink"},
	{Typ: cfgStr, Name: "server_addr", Usage: "TCP服务监听地址(任意端口)", ValStr: ":443"},
	{Typ: cfgBool, Name: "server_dtls", Usage: "开启DTLS", ValBool: false},
	{Typ: cfgStr, Name: "server_dtls_addr", Usage: "DTLS监听地址(任意端口)", ValStr: ":443"},
	{Typ: cfgStr, Name: "advertise_dtls_addr", Usage: "DTLS对外映射端口(为空则与server_dtls_addr相同)", ValStr: ""},
	{Typ: cfgStr, Name: "admin_addr", Usage: "后台服务监听地址", ValStr: ":8800"},
	{Typ: cfgBool, Name: "proxy_protocol", Usage: "TCP代理协议", ValBool: false},
	{Typ: cfgStr, Name: "db_type", Usage: "数据库类型 [sqlite3 mysql postgres]", ValStr: "sqlite3"},
	{Typ: cfgStr, Name: "db_source", Usage: "数据库source", ValStr: "./conf/anylink.db"},
	{Typ: cfgStr, Name: "cert_file", Usage: "证书文件", ValStr: "./conf/vpn_cert.pem"},
	{Typ: cfgStr, Name: "cert_key", Usage: "证书密钥", ValStr: "./conf/vpn_cert.key"},
	{Typ: cfgStr, Name: "files_path", Usage: "外部下载文件路径", ValStr: "./conf/files"},
	{Typ: cfgStr, Name: "log_path", Usage: "日志文件路径,默认标准输出", ValStr: ""},
	{Typ: cfgStr, Name: "log_level", Usage: "日志等级 [debug info warn error]", ValStr: "debug"},
	{Typ: cfgBool, Name: "http_server_log", Usage: "开启go标准库http.Server的日志", ValBool: false},
	{Typ: cfgBool, Name: "pprof", Usage: "开启pprof", ValBool: true},
	{Typ: cfgStr, Name: "issuer", Usage: "系统名称", ValStr: "XX公司VPN"},
	{Typ: cfgStr, Name: "admin_user", Usage: "管理用户名", ValStr: "admin"},
	{Typ: cfgStr, Name: "admin_pass", Usage: "管理用户密码", ValStr: defaultPwd},
	{Typ: cfgStr, Name: "admin_otp", Usage: "管理用户otp,生成命令 ./anylink tool -o", ValStr: ""},
	{Typ: cfgStr, Name: "jwt_secret", Usage: "JWT密钥", ValStr: defaultJwt},
	{Typ: cfgStr, Name: "link_mode", Usage: "虚拟网络类型[tun tap macvtap ipvtap]", ValStr: "tun"},
	{Typ: cfgStr, Name: "ipv4_master", Usage: "ipv4主网卡名称", ValStr: "eth0"},
	{Typ: cfgStr, Name: "ipv4_cidr", Usage: "ip地址网段", ValStr: "192.168.90.0/24"},
	{Typ: cfgStr, Name: "ipv4_gateway", Usage: "ipv4_gateway", ValStr: "192.168.90.1"},
	{Typ: cfgStr, Name: "ipv4_start", Usage: "IPV4开始地址", ValStr: "192.168.90.100"},
	{Typ: cfgStr, Name: "ipv4_end", Usage: "IPV4结束", ValStr: "192.168.90.200"},
	{Typ: cfgStr, Name: "default_group", Usage: "默认用户组", ValStr: "one"},
	{Typ: cfgStr, Name: "default_domain", Usage: "客户端dns的默认搜索域", ValStr: ""},

	{Typ: cfgInt, Name: "ip_lease", Usage: "IP租期(秒)", ValInt: 86400},
	{Typ: cfgInt, Name: "max_client", Usage: "最大用户连接", ValInt: 200},
	{Typ: cfgInt, Name: "max_user_client", Usage: "最大单用户连接", ValInt: 3},
	{Typ: cfgInt, Name: "cstp_keepalive", Usage: "keepalive时间(秒)", ValInt: 3},
	{Typ: cfgInt, Name: "cstp_dpd", Usage: "死链接检测时间(秒)", ValInt: 20},
	{Typ: cfgInt, Name: "mobile_keepalive", Usage: "移动端keepalive接检测时间(秒)", ValInt: 4},
	{Typ: cfgInt, Name: "mobile_dpd", Usage: "移动端死链接检测时间(秒)", ValInt: 60},
	{Typ: cfgInt, Name: "mtu", Usage: "最大传输单元MTU", ValInt: 1460},
	{Typ: cfgInt, Name: "idle_timeout", Usage: "空闲链接超时时间(秒)-超时后断开链接，0关闭此功能", ValInt: 0},
	{Typ: cfgInt, Name: "session_timeout", Usage: "session过期时间(秒)-用于断线重连，0永不过期", ValInt: 3600},
	// {Typ: cfgInt, Name: "auth_timeout", Usage: "auth_timeout", ValInt: 0},
	{Typ: cfgInt, Name: "audit_interval", Usage: "审计去重间隔(秒),-1关闭", ValInt: 600},

	{Typ: cfgBool, Name: "show_sql", Usage: "显示sql语句，用于调试", ValBool: false},
	{Typ: cfgBool, Name: "iptables_nat", Usage: "是否自动添加NAT", ValBool: true},
	{Typ: cfgBool, Name: "compression", Usage: "启用压缩", ValBool: false},
	{Typ: cfgInt, Name: "no_compress_limit", Usage: "低于及等于多少字节不压缩", ValInt: 256},

	{Typ: cfgBool, Name: "display_error", Usage: "客户端显示详细错误信息(线上环境慎开启)", ValBool: false},
	{Typ: cfgBool, Name: "exclude_export_ip", Usage: "排除出口ip路由(出口ip不加密传输)", ValBool: true},
	{Typ: cfgBool, Name: "auth_alone_otp", Usage: "登录单独验证OTP窗口", ValBool: false},
	{Typ: cfgBool, Name: "encryption_password", Usage: "用户密码是否加密保存", ValBool: false},

	{Typ: cfgBool, Name: "anti_brute_force", Usage: "是否开启防爆功能", ValBool: true},
	{Typ: cfgStr, Name: "ip_whitelist", Usage: "全局IP白名单,多个用逗号分隔，支持单IP和CIDR范围", ValStr: "192.168.90.1,172.16.0.0/24"},

	{Typ: cfgInt, Name: "max_ban_score", Usage: "单位时间内最大尝试次数，0为关闭该功能", ValInt: 5},
	{Typ: cfgInt, Name: "ban_reset_time", Usage: "设置单位时间(秒)，超过则重置计数", ValInt: 10},
	{Typ: cfgInt, Name: "lock_time", Usage: "超过最大尝试次数后的锁定时长(秒)", ValInt: 300},

	{Typ: cfgInt, Name: "max_global_user_ban_count", Usage: "全局用户单位时间内最大尝试次数，0为关闭该功能", ValInt: 20},
	{Typ: cfgInt, Name: "global_user_ban_reset_time", Usage: "全局用户设置单位时间(秒)", ValInt: 600},
	{Typ: cfgInt, Name: "global_user_lock_time", Usage: "全局用户锁定时间(秒)", ValInt: 300},

	{Typ: cfgInt, Name: "max_global_ip_ban_count", Usage: "全局IP单位时间内最大尝试次数，0为关闭该功能", ValInt: 40},
	{Typ: cfgInt, Name: "global_ip_ban_reset_time", Usage: "全局IP设置单位时间(秒)", ValInt: 1200},
	{Typ: cfgInt, Name: "global_ip_lock_time", Usage: "全局IP锁定时间(秒)", ValInt: 300},

	{Typ: cfgInt, Name: "global_lock_state_expiration_time", Usage: "全局锁定状态的保存生命周期(秒),超过则删除记录", ValInt: 3600},
}

var envs = map[string]string{}
