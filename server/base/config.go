package base

const (
	cfgStr = iota
	cfgInt
	cfgBool
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
	{Typ: cfgStr, Name: "link_addr", Usage: "vpn服务对外地址", ValStr: "vpn.xx.com"},
	{Typ: cfgStr, Name: "server_addr", Usage: "前台服务监听地址", ValStr: ":443"},
	{Typ: cfgStr, Name: "admin_addr", Usage: "后台服务监听地址", ValStr: ":8800"},
	{Typ: cfgBool, Name: "proxy_protocol", Usage: "TCP代理协议", ValBool: false},
	{Typ: cfgStr, Name: "db_file", Usage: "数据库地址", ValStr: "./conf/data.db"},
	{Typ: cfgStr, Name: "cert_file", Usage: "证书文件", ValStr: "./conf/vpn_cert.pem"},
	{Typ: cfgStr, Name: "cert_key", Usage: "证书密钥", ValStr: "./conf/vpn_cert.key"},
	{Typ: cfgStr, Name: "ui_path", Usage: "ui文件路径", ValStr: "./ui"},
	{Typ: cfgStr, Name: "files_path", Usage: "外部下载文件路径", ValStr: "./conf/files"},
	{Typ: cfgStr, Name: "log_path", Usage: "日志文件路径", ValStr: ""},
	{Typ: cfgStr, Name: "log_level", Usage: "日志等级", ValStr: "info"},
	{Typ: cfgStr, Name: "issuer", Usage: "系统名称", ValStr: "XX公司VPN"},
	{Typ: cfgStr, Name: "admin_user", Usage: "管理用户名", ValStr: "admin"},
	{Typ: cfgStr, Name: "admin_pass", Usage: "管理用户密码", ValStr: ""},
	{Typ: cfgStr, Name: "jwt_secret", Usage: "JWT密钥", ValStr: ""},
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

var envs = map[string]string{"admin_addr": "LINK_ADMIN_ADDR", "admin_pass": "LINK_ADMIN_PASS", "admin_user": "LINK_ADMIN_USER", "cert_file": "LINK_CERT_FILE", "cert_key": "LINK_CERT_KEY", "cstp_dpd": "LINK_CSTP_DPD", "cstp_keepalive": "LINK_CSTP_KEEPALIVE", "db_file": "LINK_DB_FILE", "default_group": "LINK_DEFAULT_GROUP", "files_path": "LINK_FILES_PATH", "ip_lease": "LINK_IP_LEASE", "ipv4_cidr": "LINK_IPV4_CIDR", "ipv4_end": "LINK_IPV4_END", "ipv4_gateway": "LINK_IPV4_GATEWAY", "ipv4_start": "LINK_IPV4_START", "issuer": "LINK_ISSUER", "jwt_secret": "LINK_JWT_SECRET", "link_addr": "LINK_LINK_ADDR", "link_mode": "LINK_LINK_MODE", "log_level": "LINK_LOG_LEVEL", "log_path": "LINK_LOG_PATH", "max_client": "LINK_MAX_CLIENT", "max_user_client": "LINK_MAX_USER_CLIENT", "mobile_dpd": "LINK_MOBILE_DPD", "mobile_keepalive": "LINK_MOBILE_KEEPALIVE", "proxy_protocol": "LINK_PROXY_PROTOCOL", "server_addr": "LINK_SERVER_ADDR", "session_timeout": "LINK_SESSION_TIMEOUT", "ui_path": "LINK_UI_PATH"}
