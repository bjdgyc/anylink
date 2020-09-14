package common

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

const (
	LinkModeTUN = "tun"
	LinkModeTAP = "tap"
)

var (
	ServerCfg = &ServerConfig{}
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
	ServerAddr    string `toml:"server_addr"`
	AdminAddr     string `toml:"admin_addr"`
	ProxyProtocol bool   `toml:"proxy_protocol"`
	DbFile        string `toml:"db_file"`
	CertFile      string `toml:"cert_file"`
	CertKey       string `toml:"cert_key"`
	LogLevel      string `toml:"log_level"`

	LinkMode      string   `toml:"link_mode"`    // tun tap
	Ipv4Network   string   `toml:"ipv4_network"` // 192.168.1.0
	Ipv4Netmask   string   `toml:"ipv4_netmask"` // 255.255.255.0
	Ipv4Gateway   string   `toml:"ipv4_gateway"`
	Ipv4Pool      []string `toml:"ipv4_pool"`  // Pool[0]=192.168.1.100 Pool[1]=192.168.1.200
	Include       []string `toml:"include"`    // 10.10.10.0/255.255.255.0
	Exclude       []string `toml:"exclude"`    // 192.168.5.0/255.255.255.0
	ClientDns     []string `toml:"client_dns"` // 114.114.114.114
	AllowLan      bool     `toml:"allow_lan"`  // 允许本地LAN访问vpn网络
	MaxClient     int      `toml:"max_client"`
	MaxUserClient int      `toml:"max_user_client"`

	UserGroups     []string `toml:"user_groups"`
	DefaultGroup   string   `toml:"default_group"`
	Banner         string   `toml:"banner"`   // 欢迎语
	CstpDpd        int      `toml:"cstp_dpd"` // Dead peer detection in seconds
	MobileDpd      int      `toml:"mobile_dpd"`
	CstpKeepalive  int      `toml:"cstp_keepalive"`  // in seconds
	SessionTimeout int      `toml:"session_timeout"` // in seconds
	AuthTimeout    int      `toml:"auth_timeout"`    // in seconds
}

func loadServer() {
	b, err := ioutil.ReadFile(serverFile)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(b, ServerCfg)
	if err != nil {
		panic(err)
	}

	sf, _ := filepath.Abs(serverFile)
	base := filepath.Dir(sf)

	// 转换成绝对路径
	ServerCfg.DbFile = getAbsPath(base, ServerCfg.DbFile)
	ServerCfg.CertFile = getAbsPath(base, ServerCfg.CertFile)
	ServerCfg.CertKey = getAbsPath(base, ServerCfg.CertKey)

	fmt.Printf("ServerCfg: %+v \n", ServerCfg)
}

func getAbsPath(base, cfile string) string {
	abs := filepath.IsAbs(cfile)
	if abs {
		return cfile
	}
	return filepath.Join(base, cfile)
}
