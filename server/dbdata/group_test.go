package dbdata

import (
	"testing"

	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetGroupNames(t *testing.T) {
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	// 添加 group
	g1 := Group{Name: "g1", ClientDns: []ValData{{Val: "114.114.114.114"}}}
	err := SetGroup(&g1)
	ast.Nil(err)
	g2 := Group{Name: "g2", ClientDns: []ValData{{Val: "114.114.114.114"}}}
	err = SetGroup(&g2)
	ast.Nil(err)
	g3 := Group{Name: "g3", ClientDns: []ValData{{Val: "114.114.114.114"}}}
	err = SetGroup(&g3)
	ast.Nil(err)

	authData := map[string]interface{}{
		"type": "radius",
		"radius": map[string]string{
			"addr":   "192.168.8.12:1044",
			"secret": "43214132",
		},
	}
	g4 := Group{Name: "g4", ClientDns: []ValData{{Val: "114.114.114.114"}}, Auth: authData}
	err = SetGroup(&g4)
	ast.Nil(err)
	g5 := Group{Name: "g5", ClientDns: []ValData{{Val: "114.114.114.114"}}, DsIncludeDomains: "baidu.com,163.com"}
	err = SetGroup(&g5)
	if ast.NotNil(err) {
		ast.Equal("默认路由, 不允许设置\"包含域名\", 请重新配置", err.Error())
	}
	g6 := Group{Name: "g6", ClientDns: []ValData{{Val: "114.114.114.114"}}, DsExcludeDomains: "com.cn,qq.com"}
	err = SetGroup(&g6)
	ast.Nil(err)

	authData = map[string]interface{}{
		"type": "ldap",
		"ldap": map[string]interface{}{
			"addr":         "192.168.8.12:389",
			"tls":          true,
			"bind_name":    "userfind@abc.com",
			"bind_pwd":     "afdbfdsafds",
			"base_dn":      "dc=abc,dc=com",
			"object_class": "person",
			"search_attr":  "sAMAccountName",
			"member_of":    "cn=vpn,cn=user,dc=abc,dc=com",
		},
	}
	g7 := Group{Name: "g7", ClientDns: []ValData{{Val: "114.114.114.114"}}, Auth: authData}
	err = SetGroup(&g7)
	ast.Nil(err)

	// 判断所有数据
	gAll := []string{"g1", "g2", "g3", "g4", "g5", "g6", "g7"}
	gs := GetGroupNames()
	for _, v := range gs {
		ast.Equal(true, utils.InArrStr(gAll, v))
	}

	gni := GetGroupNamesIds()
	for _, v := range gni {
		ast.NotEqual(0, v.Id)
		ast.Equal(true, utils.InArrStr(gAll, v.Name))
	}
}
