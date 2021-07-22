package dbdata

import (
	"errors"
	"time"
)

// type IpMap struct {
// 	Id        int       `json:"id" xorm:"pk autoincr not null"`
// 	IpAddr    string    `json:"ip_addr" xorm:"not null unique"`
// 	MacAddr   string    `json:"mac_addr" xorm:"not null unique"`
// 	Username  string    `json:"username"`
// 	Keep      bool      `json:"keep"` // 保留 ip-mac 绑定
// 	KeepTime  time.Time `json:"keep_time"`
// 	Note      string    `json:"note"` // 备注
// 	LastLogin time.Time `json:"last_login"`
// 	UpdatedAt time.Time `json:"updated_at"`
// }

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
