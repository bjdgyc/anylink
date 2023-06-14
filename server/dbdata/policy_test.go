package dbdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPolicy(t *testing.T) {
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	// 添加 Policy
	p1 := Policy{Username: "a1", ClientDns: []ValData{{Val: "114.114.114.114"}}, DsExcludeDomains: "baidu.com,163.com"}
	err := SetPolicy(&p1)
	ast.Nil(err)

	p2 := Policy{Username: "a2", ClientDns: []ValData{{Val: "114.114.114.114"}}, DsExcludeDomains: "com.cn,qq.com"}
	err = SetPolicy(&p2)
	ast.Nil(err)

	route := []ValData{{Val: "192.168.1.0/24"}}
	p3 := Policy{Username: "a3", ClientDns: []ValData{{Val: "114.114.114.114"}}, RouteInclude: route, DsExcludeDomains: "com.cn,qq.com"}
	err = SetPolicy(&p3)
	ast.Nil(err)
	// 判断 IpMask
	ast.Equal(p3.RouteInclude[0].IpMask, "192.168.1.0/255.255.255.0")

	route2 := []ValData{{Val: "192.168.2.0/24"}}
	p4 := Policy{Username: "a4", ClientDns: []ValData{{Val: "114.114.114.114"}}, RouteExclude: route2, DsIncludeDomains: "com.cn,qq.com"}
	err = SetPolicy(&p4)
	ast.Nil(err)
	// 判断 IpMask
	ast.Equal(p4.RouteExclude[0].IpMask, "192.168.2.0/255.255.255.0")

	// 判断所有数据
	var userPolicy *Policy
	pAll := []string{"a1", "a2", "a3", "a4"}
	for _, v := range pAll {
		userPolicy = GetPolicy(v)
		ast.NotEqual(userPolicy.Id, 0, "user policy id is zero")
	}
}
