package dbdata

import (
	"encoding/json"
	"net"
	"time"
)

const BucketGroup = "group"

type Group struct {
	Id           int
	Name         string
	RouteInclude []string
	RouteExclude []string
	AllowLan     bool
	LinkAcl      []struct {
		Action string // allow、deny
		IpNet  string
		IPNet  net.IPNet
	}
	Bandwidth int // 带宽限制
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetGroups(lastKey string, prev bool) []Group {
	res := getList(BucketUser, lastKey, prev)
	datas := make([]Group, 0)
	for _, data := range res {
		d := Group{}
		json.Unmarshal(data, &d)
		datas = append(datas, d)
	}
	return datas
}
