package dbdata

import (
	"net"
	"time"
)

const (
	Allow = "allow"
	Deny  = "deny"
)

type GroupLinkAcl struct {
	// 自上而下匹配 默认 allow * *
	Action string     `json:"action"` // allow、deny
	Val    string     `json:"val"`
	Port   uint16     `json:"port"`
	IpNet  *net.IPNet `json:"ip_net"`
	Note   string     `json:"note"`
}

type ValData struct {
	Val    string `json:"val"`
	IpMask string `json:"ip_mask"`
	Note   string `json:"note"`
}

type Group struct {
	Id           int            `json:"id" xorm:"pk autoincr not null"`
	Name         string         `json:"name" xorm:"varchar(255) not null index(idx_code) unique"`
	Note         string         `json:"note" xorm:"varchar(255)"`
	AllowLan     bool           `json:"allow_lan" xorm:"Bool"`
	ClientDns    []ValData      `json:"client_dns" xorm:"Text"`
	RouteInclude []ValData      `json:"route_include" xorm:"Text"`
	RouteExclude []ValData      `json:"route_exclude" xorm:"Text"`
	LinkAcl      []GroupLinkAcl `json:"link_acl" xorm:"Text"`
	Bandwidth    int            `json:"bandwidth" xorm:"Int"` // 带宽限制
	Status       int8           `json:"status" xorm:"Int"`    // 1正常
	CreatedAt    time.Time      `json:"created_at" xorm:"DateTime"`
	UpdatedAt    time.Time      `json:"updated_at" xorm:"DateTime"`
}

type User struct {
	Id       int    `json:"id" xorm:"pk autoincr not null"`
	Username string `json:"username" xorm:"varchar(255) not null index(idx_code) unique"`
	Nickname string `json:"nickname" xorm:"varchar(255)"`
	Email    string `json:"email" xorm:"varchar(255)"`
	// Password  string    `json:"password"`
	PinCode    string    `json:"pin_code" xorm:"varchar(255)"`
	OtpSecret  string    `json:"otp_secret" xorm:"varchar(255)"`
	DisableOtp bool      `json:"disable_otp" xorm:"Bool"` // 禁用otp
	Groups     []string  `json:"groups" xorm:"Text"`
	Status     int8      `json:"status" xorm:"Int"` // 1正常
	SendEmail  bool      `json:"send_email" xorm:"Bool"`
	CreatedAt  time.Time `json:"created_at" xorm:"DateTime"`
	UpdatedAt  time.Time `json:"updated_at" xorm:"DateTime"`
}

type SettingSmtp struct {
	Id         int    `json:"id" xorm:"pk autoincr not null"`
	Host       string `json:"host" xorm:"varchar(255) not null index(idx_code)"`
	Port       int    `json:"port" xorm:"Int"`
	Username   string `json:"username" xorm:"varchar(255)"`
	Password   string `json:"password" xorm:"varchar(255)"`
	From       string `json:"from" xorm:"varchar(255)"`
	Encryption string `json:"encryption" xorm:"varchar(255)"`
}

type SettingOther struct {
	Id          int    `json:"id" xorm:"pk autoincr not null"`
	LinkAddr    string `json:"link_addr" xorm:"varchar(255)"`
	Banner      string `json:"banner" xorm:"varchar(255)"`
	AccountMail string `json:"account_mail" xorm:"varchar(2048)"`
}

type IpMap struct {
	Id        int       `json:"id" xorm:"pk autoincr not null"`
	IpAddr    net.IP    `json:"ip_addr" xorm:"Text"`
	MacAddr   string    `json:"mac_addr" xorm:"varchar(32) index(idx_macaddr)"`
	Username  string    `json:"username" xorm:"varchar(255)"`
	Keep      bool      `json:"keep" xorm:"Bool"` // 保留 ip-mac 绑定
	KeepTime  time.Time `json:"keep_time" xorm:"DateTime"`
	Note      string    `json:"note" xorm:"varchar(255)"` // 备注
	LastLogin time.Time `json:"last_login" xorm:"DateTime"`
	UpdatedAt time.Time `json:"updated_at" xorm:"DateTime"`
}
