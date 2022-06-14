package dbdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlzd/gotp"
)

func TestCheckUser(t *testing.T) {
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	group := "group1"

	// 添加一个组
	dns := []ValData{{Val: "114.114.114.114"}}
	route := []ValData{{Val: "192.168.1.1/24"}}
	g := Group{Name: group, Status: 1, ClientDns: dns, RouteInclude: route}
	err := SetGroup(&g)
	ast.Nil(err)
	// 判断 IpMask
	ast.Equal(g.RouteInclude[0].IpMask, "192.168.1.1/255.255.255.0")

	// 添加一个用户
	u := User{Username: "aaa", Groups: []string{group}, Status: 1}
	err = SetUser(&u)
	ast.Nil(err)

	// 验证 PinCode + OtpSecret
	totp := gotp.NewDefaultTOTP(u.OtpSecret)
	secret := totp.Now()
	err = CheckUser("aaa", u.PinCode+secret, group)
	ast.Nil(err)

	// 单独验证密码
	u.DisableOtp = true
	_ = SetUser(&u)
	err = CheckUser("aaa", u.PinCode, group)
	ast.Nil(err)

	// 添加一个radius组
	group2 := "group2"
	authData := map[string]interface{}{
		"type": "radius",
		"radius": map[string]string{
			"addr":   "192.168.1.12:1044",
			"secret": "43214132",
		},
	}
	g2 := Group{Name: group2, Status: 1, ClientDns: dns, RouteInclude: route, Auth: authData}
	err = SetGroup(&g2)
	ast.Nil(err)
	err = CheckUser("aaa", "bbbbbbb", group2)
	if ast.NotNil(err) {
		ast.Equal("aaa Radius服务器连接异常, 请检测服务器和端口", err.Error())

	}
	// 添加用户策略
	dns2 := []ValData{{Val: "8.8.8.8"}}
	route2 := []ValData{{Val: "192.168.2.1/24"}}
	p1 := Policy{Username: "aaa", Status: 1, ClientDns: dns2, RouteInclude: route2}
	err = SetPolicy(&p1)
	ast.Nil(err)
	err = CheckUser("aaa", u.PinCode, group)
	ast.Nil(err)
}
