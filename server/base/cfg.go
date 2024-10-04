package base

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

const (
	LinkModeTUN     = "tun"
	LinkModeTAP     = "tap"
	LinkModeMacvtap = "macvtap"
	LinkModeIpvtap  = "ipvtap"
)

var (
	Cfg = &ServerConfig{}
)

// # ReKey time (in seconds)
// rekey-time = 172800
// # ReKey method
// # Valid options: ssl, new-tunnel
// #  ssl: Will perform an efficient rehandshake on the channel allowing
// #       a seamless connection during rekey.
// #  new-tunnel: Will instruct the client to discard and re-establish the channel.
// #       Use this option only if the connecting clients have issues with the ssl
// #       option.
// rekey-method = ssl

type ServerConfig struct {
	// LinkAddr      string `json:"link_addr"`
	Conf           string `json:"conf"`
	Profile        string `json:"profile"`
	ProfileName    string `json:"profile_name"`
	ServerAddr     string `json:"server_addr"`
	ServerDTLSAddr string `json:"server_dtls_addr"`
	ServerDTLS     bool   `json:"server_dtls"`
	AdminAddr      string `json:"admin_addr"`
	ProxyProtocol  bool   `json:"proxy_protocol"`
	DbType         string `json:"db_type"`
	DbSource       string `json:"db_source"`
	CertFile       string `json:"cert_file"`
	CertKey        string `json:"cert_key"`
	FilesPath      string `json:"files_path"`
	LogPath        string `json:"log_path"`
	LogLevel       string `json:"log_level"`
	HttpServerLog  bool   `json:"http_server_log"`
	Pprof          bool   `json:"pprof"`
	Issuer         string `json:"issuer"`
	AdminUser      string `json:"admin_user"`
	AdminPass      string `json:"admin_pass"`
	AdminOtp       string `json:"admin_otp"`
	JwtSecret      string `json:"jwt_secret"`

	LinkMode    string `json:"link_mode"`    // tun tap macvtap ipvtap
	Ipv4Master  string `json:"ipv4_master"`  // eth0
	Ipv4CIDR    string `json:"ipv4_cidr"`    // 192.168.10.0/24
	Ipv4Gateway string `json:"ipv4_gateway"` // 192.168.10.1
	Ipv4Start   string `json:"ipv4_start"`   // 192.168.10.100
	Ipv4End     string `json:"ipv4_end"`     // 192.168.10.200
	IpLease     int    `json:"ip_lease"`

	MaxClient       int    `json:"max_client"`
	MaxUserClient   int    `json:"max_user_client"`
	DefaultGroup    string `json:"default_group"`
	CstpKeepalive   int    `json:"cstp_keepalive"` // in seconds
	CstpDpd         int    `json:"cstp_dpd"`       // Dead peer detection in seconds
	MobileKeepalive int    `json:"mobile_keepalive"`
	MobileDpd       int    `json:"mobile_dpd"`
	Mtu             int    `json:"mtu"`
	DefaultDomain   string `json:"default_domain"`

	IdleTimeout    int `json:"idle_timeout"`    // in seconds
	SessionTimeout int `json:"session_timeout"` // in seconds
	// AuthTimeout    int `json:"auth_timeout"`    // in seconds
	AuditInterval int `json:"audit_interval"` // in seconds

	ShowSQL         bool `json:"show_sql"` // bool
	IptablesNat     bool `json:"iptables_nat"`
	Compression     bool `json:"compression"`       // bool
	NoCompressLimit int  `json:"no_compress_limit"` // int

	DisplayError    bool `json:"display_error"`
	ExcludeExportIp bool `json:"exclude_export_ip"`

	AntiBruteForce bool `json:"anti_brute_force"`

	MaxBanCount  int `json:"max_ban_score"`
	BanResetTime int `json:"ban_reset_time"`
	LockTime     int `json:"lock_time"`

	MaxGlobalUserBanCount  int `json:"max_global_user_ban_count"`
	GlobalUserBanResetTime int `json:"global_user_ban_reset_time"`
	GlobalUserLockTime     int `json:"global_user_lock_time"`

	MaxGlobalIPBanCount  int `json:"max_global_ip_ban_count"`
	GlobalIPBanResetTime int `json:"global_ip_ban_reset_time"`
	GlobalIPLockTime     int `json:"global_ip_lock_time"`

	GlobalLockStateExpirationTime int `json:"global_lock_state_expiration_time"`
}

func initServerCfg() {

	// TODO 取消绝对地址转换
	// sf, _ := filepath.Abs(cfgFile)
	// base := filepath.Dir(sf)

	// 转换成绝对路径
	// Cfg.DbFile = getAbsPath(base, Cfg.DbFile)
	// Cfg.CertFile = getAbsPath(base, Cfg.CertFile)
	// Cfg.CertKey = getAbsPath(base, Cfg.CertKey)
	// Cfg.UiPath = getAbsPath(base, Cfg.UiPath)
	// Cfg.FilesPath = getAbsPath(base, Cfg.FilesPath)
	// Cfg.LogPath = getAbsPath(base, Cfg.LogPath)

	if Cfg.AdminPass == defaultPwd {
		fmt.Fprintln(os.Stderr, "=== 使用默认的admin_pass有安全风险，请设置新的admin_pass ===")
	}

	if Cfg.JwtSecret == defaultJwt {
		fmt.Fprintln(os.Stderr, "=== 使用默认的jwt_secret有安全风险，请设置新的jwt_secret ===")
	}

	fmt.Printf("ServerCfg: %+v \n", Cfg)
}

func getAbsPath(base, cfile string) string {
	if cfile == "" {
		return ""
	}

	abs := filepath.IsAbs(cfile)
	if abs {
		return cfile
	}
	return filepath.Join(base, cfile)
}

func initCfg() {
	ref := reflect.ValueOf(Cfg)
	s := ref.Elem()

	typ := s.Type()
	numFields := s.NumField()
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		value := s.Field(i)
		tag := field.Tag.Get("json")

		for _, v := range configs {
			if v.Name == tag {
				if v.Typ == cfgStr {
					value.SetString(linkViper.GetString(v.Name))
				}
				if v.Typ == cfgInt {
					value.SetInt(int64(linkViper.GetInt(v.Name)))
				}
				if v.Typ == cfgBool {
					value.SetBool(linkViper.GetBool(v.Name))
				}
			}
		}
	}

	initServerCfg()
}

type SCfg struct {
	Name string      `json:"name"`
	Env  string      `json:"env"`
	Info string      `json:"info"`
	Data interface{} `json:"data"`
	Val  interface{} `json:"default"`
}

func ServerCfg2Slice() []SCfg {
	ref := reflect.ValueOf(Cfg)
	s := ref.Elem()

	var datas []SCfg

	typ := s.Type()
	numFields := s.NumField()
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		value := s.Field(i)
		tag := field.Tag.Get("json")
		usage, env, val := getUsageEnv(tag)

		datas = append(datas, SCfg{Name: tag, Env: env, Info: usage, Data: value.Interface(), Val: val})
	}

	return datas
}

func getUsageEnv(name string) (usage, env string, val interface{}) {
	for _, v := range configs {
		if v.Name == name {
			usage = v.Usage
			if v.Typ == cfgStr {
				val = v.ValStr
			}
			if v.Typ == cfgInt {
				val = v.ValInt
			}
			if v.Typ == cfgBool {
				val = v.ValBool
			}
		}
	}

	if e, ok := envs[name]; ok {
		env = e
	}

	return
}
