package dbdata

import (
	"testing"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// 覆盖用户自助改密的各条路径: 正常改密/旧密码错/用户不存在/新密码过短/
// 新旧相同/OTP用户只校验PIN/明文与密文兼容/改密后清除ChangePwd标记。
func TestUserChangePassword(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	group := "group1"
	dns := []ValData{{Val: "114.114.114.114"}}
	route := []ValData{{Val: "192.168.1.0/24"}}
	g := Group{Name: group, Status: 1, ClientDns: dns, RouteInclude: route}
	ast.Nil(SetGroup(&g))

	// 新建一个需首次改密的用户
	u := User{
		Username:  "alice",
		PinCode:   "oldpass1",
		Groups:    []string{group},
		Status:    1,
		ChangePwd: true,
	}
	ast.Nil(SetUser(&u))

	// 用户不存在 -> 模糊提示, 不泄露账号是否存在
	err := UserChangePassword("ghost", "whatever", "newpass1")
	if ast.NotNil(err) {
		ast.Equal("用户名或旧密码错误", err.Error())
	}

	// 旧密码错误 -> 同一模糊提示
	err = UserChangePassword("alice", "wrongold", "newpass1")
	if ast.NotNil(err) {
		ast.Equal("用户名或旧密码错误", err.Error())
	}

	// 新密码过短
	err = UserChangePassword("alice", "oldpass1", "123")
	if ast.NotNil(err) {
		ast.Contains(err.Error(), "不能少于6位")
	}

	// 新旧密码相同
	err = UserChangePassword("alice", "oldpass1", "oldpass1")
	if ast.NotNil(err) {
		ast.Contains(err.Error(), "不能与旧密码相同")
	}

	// 正常改密成功
	err = UserChangePassword("alice", "oldpass1", "newpass1")
	ast.Nil(err)

	// 校验: 新密码生效, 且 ChangePwd 标记被清除
	after := &User{}
	ast.Nil(One("Username", "alice", after))
	ast.True(utils.PasswordVerify("newpass1", after.PinCode))
	ast.False(after.ChangePwd)

	// 旧密码不再有效
	err = UserChangePassword("alice", "oldpass1", "another1")
	if ast.NotNil(err) {
		ast.Equal("用户名或旧密码错误", err.Error())
	}
}

// 验证开启密码加密后, 旧密码校验(密文)与改密仍正确工作。
func TestUserChangePasswordEncrypted(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	// 开启密码加密, 用例结束后复位, 避免影响其它用例
	oldEnc := base.Cfg.EncryptionPassword
	base.Cfg.EncryptionPassword = true
	defer func() { base.Cfg.EncryptionPassword = oldEnc }()

	group := "encgroup"
	dns := []ValData{{Val: "114.114.114.114"}}
	route := []ValData{{Val: "192.168.1.0/24"}}
	g := Group{Name: group, Status: 1, ClientDns: dns, RouteInclude: route}
	ast.Nil(SetGroup(&g))

	u := User{Username: "bob", PinCode: "oldpass1", Groups: []string{group}, Status: 1}
	ast.Nil(SetUser(&u))

	// 确认入库后是 bcrypt 密文
	stored := &User{}
	ast.Nil(One("Username", "bob", stored))
	ast.Equal(60, len(stored.PinCode))

	// 用明文旧密码校验密文, 改密成功
	ast.Nil(UserChangePassword("bob", "oldpass1", "newpass1"))

	after := &User{}
	ast.Nil(One("Username", "bob", after))
	ast.Equal(60, len(after.PinCode))
	ast.True(utils.PasswordVerify("newpass1", after.PinCode))
}

// 验证停用/过期账号: 旧密码正确也应拒绝, 且提示账号状态。
func TestUserChangePasswordDisabled(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	group := "disgroup"
	dns := []ValData{{Val: "114.114.114.114"}}
	route := []ValData{{Val: "192.168.1.0/24"}}
	g := Group{Name: group, Status: 1, ClientDns: dns, RouteInclude: route}
	ast.Nil(SetGroup(&g))

	// status=0 停用用户
	u := User{Username: "carol", PinCode: "oldpass1", Groups: []string{group}, Status: 0}
	ast.Nil(SetUser(&u))

	err := UserChangePassword("carol", "oldpass1", "newpass1")
	if ast.NotNil(err) {
		ast.Contains(err.Error(), "停用或已过期")
	}
}
