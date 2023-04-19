package dbdata

import (
	"errors"
	"time"
)

type IpMap struct {
	Id        int       `json:"id" xorm:"pk autoincr not null"`
	IpAddr    string    `json:"ip_addr" xorm:"varchar(32) not null unique"`
	MacAddr   string    `json:"mac_addr" xorm:"varchar(32) not null unique"`
	UniqueMac bool      `json:"unique_mac" xorm:"Bool index"`
	Username  string    `json:"username" xorm:"varchar(60)"`
	Keep      bool      `json:"keep" xorm:"Bool"` // 保留 ip-mac 绑定
	KeepTime  time.Time `json:"keep_time" xorm:"DateTime"`
	Note      string    `json:"note" xorm:"varchar(255)"` // 备注
	LastLogin time.Time `json:"last_login" xorm:"DateTime"`
	UpdatedAt time.Time `json:"updated_at" xorm:"DateTime updated"`
}

func SetIpMap(v *IpMap) error {
	var err error

	if len(v.IpAddr) < 4 || len(v.MacAddr) < 6 {
		return errors.New("IP或MAC错误")
	}

	v.UpdatedAt = time.Now()
	if v.Id > 0 {
		err = Set(v)
	} else {
		err = Add(v)
	}
	return err
}
