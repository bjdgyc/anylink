package dbdata

import (
	"encoding/json"
	"time"
)

type Group struct {
	Id               int                    `json:"id" xorm:"pk autoincr not null"`
	Name             string                 `json:"name" xorm:"varchar(60) not null unique"`
	Note             string                 `json:"note" xorm:"varchar(255)"`
	AllowLan         bool                   `json:"allow_lan" xorm:"Bool"`
	ClientDns        []ValData              `json:"client_dns" xorm:"Text"`
	SplitDns         []ValData              `json:"split_dns" xorm:"Text"`
	RouteInclude     []ValData              `json:"route_include" xorm:"Text"`
	RouteExclude     []ValData              `json:"route_exclude" xorm:"Text"`
	DsExcludeDomains string                 `json:"ds_exclude_domains" xorm:"Text"`
	DsIncludeDomains string                 `json:"ds_include_domains" xorm:"Text"`
	LinkAcl          []GroupLinkAcl         `json:"link_acl" xorm:"Text"`
	Bandwidth        int                    `json:"bandwidth" xorm:"Int"`                           // 带宽限制
	Auth             map[string]interface{} `json:"auth" xorm:"not null default '{}' varchar(500)"` // 认证方式
	Status           int8                   `json:"status" xorm:"Int"`                              // 1正常
	CreatedAt        time.Time              `json:"created_at" xorm:"DateTime created"`
	UpdatedAt        time.Time              `json:"updated_at" xorm:"DateTime updated"`
}

type User struct {
	Id       int    `json:"id" xorm:"pk autoincr not null"`
	Username string `json:"username" xorm:"varchar(60) not null unique"`
	Nickname string `json:"nickname" xorm:"varchar(255)"`
	Email    string `json:"email" xorm:"varchar(255)"`
	// Password  string    `json:"password"`
	PinCode    string     `json:"pin_code" xorm:"varchar(64)"`
	LimitTime  *time.Time `json:"limittime,omitempty" xorm:"Datetime limittime"` // 值为null时，前端不显示
	OtpSecret  string     `json:"otp_secret" xorm:"varchar(255)"`
	DisableOtp bool       `json:"disable_otp" xorm:"Bool"` // 禁用otp
	Groups     []string   `json:"groups" xorm:"Text"`
	Status     int8       `json:"status" xorm:"Int"` // 1正常
	SendEmail  bool       `json:"send_email" xorm:"Bool"`
	CreatedAt  time.Time  `json:"created_at" xorm:"DateTime created"`
	UpdatedAt  time.Time  `json:"updated_at" xorm:"DateTime updated"`
}

type UserActLog struct {
	Id              int       `json:"id" xorm:"pk autoincr not null"`
	Username        string    `json:"username" xorm:"varchar(60)"`
	GroupName       string    `json:"group_name" xorm:"varchar(60)"`
	IpAddr          string    `json:"ip_addr" xorm:"varchar(32)"`
	RemoteAddr      string    `json:"remote_addr" xorm:"varchar(42)"`
	Os              uint8     `json:"os" xorm:"not null default 0 Int"`
	Client          uint8     `json:"client" xorm:"not null default 0 Int"`
	Version         string    `json:"version" xorm:"varchar(15)"`
	DeviceType      string    `json:"device_type" xorm:"varchar(128) not null default ''"`
	PlatformVersion string    `json:"platform_version" xorm:"varchar(128) not null default ''"`
	Status          uint8     `json:"status" xorm:"not null default 0 Int"`
	Info            string    `json:"info" xorm:"varchar(255) not null default ''"` // 详情
	CreatedAt       time.Time `json:"created_at" xorm:"DateTime created"`
}

type Setting struct {
	Id        int             `json:"id" xorm:"pk autoincr not null"`
	Name      string          `json:"name" xorm:"varchar(60) not null unique"`
	Data      json.RawMessage `json:"data" xorm:"Text"`
	UpdatedAt time.Time       `json:"updated_at" xorm:"DateTime updated"`
}

type AccessAudit struct {
	Id          int       `json:"id" xorm:"pk autoincr not null"`
	Username    string    `json:"username" xorm:"varchar(60) not null"`
	Protocol    uint8     `json:"protocol" xorm:"Int not null"`
	Src         string    `json:"src" xorm:"varchar(60) not null"`
	SrcPort     uint16    `json:"src_port" xorm:"Int not null"`
	Dst         string    `json:"dst" xorm:"varchar(60) not null"`
	DstPort     uint16    `json:"dst_port" xorm:"Int not null"`
	AccessProto uint8     `json:"access_proto" xorm:"Int default 0"`            // 访问协议
	Info        string    `json:"info" xorm:"varchar(255) not null default ''"` // 详情
	CreatedAt   time.Time `json:"created_at" xorm:"DateTime"`
}

type Policy struct {
	Id               int       `json:"id" xorm:"pk autoincr not null"`
	Username         string    `json:"username" xorm:"varchar(60) not null unique"`
	AllowLan         bool      `json:"allow_lan" xorm:"Bool"`
	ClientDns        []ValData `json:"client_dns" xorm:"Text"`
	RouteInclude     []ValData `json:"route_include" xorm:"Text"`
	RouteExclude     []ValData `json:"route_exclude" xorm:"Text"`
	DsExcludeDomains string    `json:"ds_exclude_domains" xorm:"Text"`
	DsIncludeDomains string    `json:"ds_include_domains" xorm:"Text"`
	Status           int8      `json:"status" xorm:"Int"` // 1正常 0 禁用
	CreatedAt        time.Time `json:"created_at" xorm:"DateTime created"`
	UpdatedAt        time.Time `json:"updated_at" xorm:"DateTime updated"`
}

type StatsOnline struct {
	Id        int       `json:"id" xorm:"pk autoincr not null"`
	Num       int       `json:"num" xorm:"Int"`
	NumGroups string    `json:"num_groups" xorm:"varchar(500) not null"`
	CreatedAt time.Time `json:"created_at" xorm:"DateTime created index"`
}

type StatsNetwork struct {
	Id         int       `json:"id" xorm:"pk autoincr not null"`
	Up         uint32    `json:"up" xorm:"Int"`
	Down       uint32    `json:"down" xorm:"Int"`
	UpGroups   string    `json:"up_groups" xorm:"varchar(500) not null"`
	DownGroups string    `json:"down_groups" xorm:"varchar(500) not null"`
	CreatedAt  time.Time `json:"created_at" xorm:"DateTime created index"`
}

type StatsCpu struct {
	Id        int       `json:"id" xorm:"pk autoincr not null"`
	Percent   float64   `json:"percent" xorm:"Float"`
	CreatedAt time.Time `json:"created_at" xorm:"DateTime created index"`
}

type StatsMem struct {
	Id        int       `json:"id" xorm:"pk autoincr not null"`
	Percent   float64   `json:"percent" xorm:"Float"`
	CreatedAt time.Time `json:"created_at" xorm:"DateTime created index"`
}
