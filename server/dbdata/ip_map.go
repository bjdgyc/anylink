package dbdata

import (
	"net"
	"time"
)

type IpMap struct {
	Id        int       `json:"id" storm:"id,increment"`
	IpAddr    net.IP    `json:"ip_addr" storm:"unique"`
	MacAddr   string    `json:"mac_addr" storm:"unique"`
	Username  string    `json:"username"`
	Keep      bool      `json:"keep"` // 保留 ip-mac 绑定
	KeepTime  time.Time `json:"keep_time"`
	Note      string    `json:"note"` // 备注
	LastLogin time.Time `json:"last_login"`
	UpdatedAt time.Time `json:"updated_at"`
}
