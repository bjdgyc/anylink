package dbdata

import (
	"crypto/tls"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/spf13/viper"
	"net"
	"strconv"
	"testing"
	"time"
)

type AuthTestLdap struct {
	Addr        string `json:"addr"`
	Tls         bool   `json:"tls"`
	BindName    string `json:"bind_name"`
	BindPwd     string `json:"bind_pwd"`
	BaseDn      string `json:"base_dn"`
	ObjectClass string `json:"object_class"`
	SearchAttr  string `json:"search_attr"`
	MemberOf    string `json:"member_of"`
}

func TestCheckLdapUserAuth(t *testing.T) {

	v := viper.New()
	v.SetConfigFile("../conf/server.toml")
	if err := v.ReadInConfig(); err != nil {
		panic("config file err:" + err.Error())
	}

	user, pwd, ldapAdminUser := v.Get("ldap_user").(string), v.Get("ldap_pass").(string), v.Get("ldap_admin_user").(string)
	addr, baseDN := v.Get("ldap_server").(string), v.Get("ldap_base_dn").(string)
	pl := len(pwd)

	if user == "" || pl < 1 {
		t.Errorf("%s %s", user, "密码错误")
	}

	// 检测服务器和端口的可用性
	con, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		t.Errorf("%s %s", user, "LDAP服务器连接异常, 请检测服务器和端口")
	}
	defer con.Close()

	// 连接LDAP
	l, err := ldap.Dial("tcp", addr)
	if err != nil {
		t.Errorf("LDAP连接失败 %s %s", addr, err.Error())
	}
	defer l.Close()

	var auth AuthTestLdap
	if auth.Tls {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			t.Errorf("%s LDAP TLS连接失败 %s", user, err.Error())
		}
	}

	err = l.Bind(user, pwd)
	if err != nil {
		t.Errorf("%s LDAP 管理员 DN或密码填写有误 %s", user, err.Error())
	}

	if auth.ObjectClass == "" {
		auth.ObjectClass = "person"
	}

	// 普通用户验证
	user = "test"
	searchAttr := fmt.Sprintf("(&(objectClass=person)(sAMAccountName=%s))", user)
	fmt.Println("searchAttr:", searchAttr)

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 3, false,
		fmt.Sprintf("(&%s)", searchAttr),
		[]string{},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		t.Logf("%s LDAP 查询失败 %s %s %s", user, auth.BaseDn, searchAttr, err.Error())
	}

	//验证密码和动态口令
	userDN := sr.Entries[0].DN
	fmt.Println("userDN: ", userDN)

	// 管理员用户不需要 otp 认证，或可以设置为固定的 otp，可根据自身情况调整
	if user == ldapAdminUser {
		pinCode := pwd
		err = l.Bind(userDN, pinCode)
		if err != nil {
			t.Logf("LDAP 登入失败，请检查登入的账号 [%s] 或密码 [%v], err=[%v]", userDN, pinCode, err.Error())
		}
	} else {

		pwd = "TEstestS#23$331239"
		pl = len(pwd)
		pinCode := pwd[:pl-6]
		otp := pwd[pl-6:]

		err = l.Bind(userDN, pinCode)
		if err != nil {
			t.Errorf("LDAP 登入失败，请检查登入的账号 [%s] 或密码 [%v], err=[%v]", userDN, pinCode, err.Error())
		} else {

			ot, err := strconv.Atoi(otp)

			otpAuthRes, err := ValidateUserOtp(user, ot)
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println("otpAuthRes: ", otpAuthRes)
		}

		fmt.Println("otp auth stop")
	}
}
