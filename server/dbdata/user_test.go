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
}
