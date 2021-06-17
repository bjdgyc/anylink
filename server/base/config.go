package base

const (
	cfgStr = iota
	cfgInt
	cfgBool

	defaultJwt = "abcdef.0123456789.abcdef"
)

type config struct {
	Typ     int
	Name    string
	Usage   string
	ValStr  string
	ValInt  int
	ValBool bool
}

var configs = []config{
	{Typ: cfgStr, Name: "server_addr", Usage: "服务监听地址", ValStr: ":443"},
	{Typ: cfgBool, Name: "server_dtls", Usage: "开启DTLS", ValBool: false},
	{Typ: cfgStr, Name: "server_dtls_addr", Usage: "DTLS监听地址", ValStr: ":4433"},
	{Typ: cfgStr, Name: "admin_addr", Usage: "后台服务监听地址", ValStr: ":8800"},
	{Typ: cfgBool, Name: "proxy_protocol", Usage: "TCP代理协议", ValBool: false},
	{Typ: cfgStr, Name: "db_file", Usage: "数据库地址", ValStr: "./data.db"},
	{Typ: cfgStr, Name: "cert_file", Usage: "证书文件", ValStr: "./vpn_cert.pem"},
	{Typ: cfgStr, Name: "cert_key", Usage: "证书密钥", ValStr: "./vpn_cert.key"},
	{Typ: cfgStr, Name: "files_path", Usage: "外部下载文件路径", ValStr: "./files"},
	{Typ: cfgStr, Name: "log_path", Usage: "日志文件路径", ValStr: ""},
	{Typ: cfgStr, Name: "log_level", Usage: "日志等级", ValStr: "info"},
	{Typ: cfgBool, Name: "pprof", Usage: "开启pprof", ValBool: false},
	{Typ: cfgStr, Name: "issuer", Usage: "系统名称", ValStr: "XX公司VPN"},
	{Typ: cfgStr, Name: "admin_user", Usage: "管理用户名", ValStr: "admin"},
	{Typ: cfgStr, Name: "admin_pass", Usage: "管理用户密码", ValStr: "$2a$10$UQ7C.EoPifDeJh6d8.31TeSPQU7hM/NOM2nixmBucJpAuXDQNqNke"},
	{Typ: cfgStr, Name: "jwt_secret", Usage: "JWT密钥", ValStr: defaultJwt},
	{Typ: cfgStr, Name: "link_mode", Usage: "虚拟网络类型", ValStr: "tun"},
	{Typ: cfgStr, Name: "ipv4_cidr", Usage: "ip地址网段", ValStr: "192.168.10.0/24"},
	{Typ: cfgStr, Name: "ipv4_gateway", Usage: "ipv4_gateway", ValStr: "192.168.10.1"},
	{Typ: cfgStr, Name: "ipv4_start", Usage: "IPV4开始地址", ValStr: "192.168.10.100"},
	{Typ: cfgStr, Name: "ipv4_end", Usage: "IPV4结束", ValStr: "192.168.10.200"},
	{Typ: cfgStr, Name: "default_group", Usage: "默认用户组", ValStr: "one"},

	{Typ: cfgInt, Name: "ip_lease", Usage: "IP租期(秒)", ValInt: 1209600},
	{Typ: cfgInt, Name: "max_client", Usage: "最大用户连接", ValInt: 100},
	{Typ: cfgInt, Name: "max_user_client", Usage: "最大单用户连接", ValInt: 3},
	{Typ: cfgInt, Name: "cstp_keepalive", Usage: "keepalive时间(秒)", ValInt: 20},
	{Typ: cfgInt, Name: "cstp_dpd", Usage: "死链接检测时间(秒)", ValInt: 30},
	{Typ: cfgInt, Name: "mobile_keepalive", Usage: "移动端keepalive接检测时间(秒)", ValInt: 50},
	{Typ: cfgInt, Name: "mobile_dpd", Usage: "移动端死链接检测时间(秒)", ValInt: 60},
	{Typ: cfgInt, Name: "session_timeout", Usage: "session过期时间(秒)", ValInt: 3600},
	// {Typ: cfgInt, Name: "auth_timeout", Usage: "auth_timeout", ValInt: 0},
}

var envs = map[string]string{}
