package dbdata

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"time"

	"github.com/go-ldap/ldap"
)

type AuthLdap struct {
	Addr       string `json:"addr"`
	Tls        bool   `json:"tls"`
	BindName   string `json:"bind_name"`
	BindPwd    string `json:"bind_pwd"`
	BaseDn     string `json:"base_dn"`
	SearchAttr string `json:"search_attr"`
	MemberOf   string `json:"member_of"`
}

func init() {
	authRegistry["ldap"] = reflect.TypeOf(AuthLdap{})
}

func (auth AuthLdap) checkData(authData map[string]interface{}) error {
	authType := authData["type"].(string)
	bodyBytes, err := json.Marshal(authData[authType])
	if err != nil {
		return errors.New("LDAP配置填写有误")
	}
	json.Unmarshal(bodyBytes, &auth)
	// 支持域名和IP, 必须填写端口
	if !ValidateIpPort(auth.Addr) && !ValidateDomainPort(auth.Addr) {
		return errors.New("LDAP的服务器地址(含端口)填写有误")
	}
	if auth.BindName == "" {
		return errors.New("LDAP的用户查询账号不能为空")
	}
	if auth.BindPwd == "" {
		return errors.New("LDAP的用户查询密码不能为空")
	}
	if auth.BaseDn == "" || !ValidateDN(auth.BaseDn) {
		return errors.New("LDAP的BaseName填写有误")
	}
	if auth.SearchAttr == "" {
		return errors.New("LDAP的搜索属性不能为空")
	}
	if auth.MemberOf != "" && !ValidateDN(auth.MemberOf) {
		return errors.New("LDAP的绑定DN填写有误")
	}
	return nil
}

func (auth AuthLdap) checkUser(name, pwd string, g *Group) error {
	pl := len(pwd)
	if name == "" || pl < 1 {
		return fmt.Errorf("%s %s", name, "密码错误")
	}
	authType := g.Auth["type"].(string)
	if _, ok := g.Auth[authType]; !ok {
		return fmt.Errorf("%s %s", name, "LDAP的ldap值不存在")
	}
	bodyBytes, err := json.Marshal(g.Auth[authType])
	if err != nil {
		return fmt.Errorf("%s %s", name, "LDAP Marshal出现错误")
	}
	err = json.Unmarshal(bodyBytes, &auth)
	if err != nil {
		return fmt.Errorf("%s %s", name, "LDAP Unmarshal出现错误")
	}
	// 检测服务器和端口的可用性
	_, err = net.DialTimeout("tcp", auth.Addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("%s %s", name, "LDAP服务器连接异常, 请检测服务器和端口")
	}
	// 连接LDAP
	l, err := ldap.Dial("tcp", auth.Addr)
	if err != nil {
		return fmt.Errorf("LDAP连接失败 %s %s", auth.Addr, err.Error())
	}
	defer l.Close()
	if auth.Tls {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return fmt.Errorf("%s LDAP TLS连接失败 %s", name, err.Error())
		}
	}
	err = l.Bind(auth.BindName, auth.BindPwd)
	if err != nil {
		return fmt.Errorf("%s LDAP 查询用户的账密有误,请重新检查 %s", name, err.Error())
	}
	filterAttr := "(objectClass=person)"
	filterAttr += "(" + auth.SearchAttr + "=" + name + ")"
	if auth.MemberOf != "" {
		filterAttr += "(memberOf:=" + auth.MemberOf + ")"
	}
	searchRequest := ldap.NewSearchRequest(
		auth.BaseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 3, false,
		fmt.Sprintf("(&%s)", filterAttr),
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("%s LDAP 查询失败 %s %s %s", name, auth.BaseDn, filterAttr, err.Error())
	}
	if len(sr.Entries) != 1 {
		if len(sr.Entries) == 0 {
			return fmt.Errorf("LDAP 找不到 %s 用户, 请检查用户或LDAP配置参数", name)
		}
		return fmt.Errorf("LDAP发现 %s 用户，存在多个账号", name)
	}
	userDN := sr.Entries[0].DN
	fmt.Println(userDN)
	err = l.Bind(userDN, pwd)
	if err != nil {
		return fmt.Errorf("%s LDAP 登入失败，请检查登入的账号或密码 %s", name, err.Error())
	}
	return nil
}

func ValidateDomainPort(addr string) bool {
	re := regexp.MustCompile(`^([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.)+[A-Za-z]{2,18}\:([0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-5]{2}[0-3][0-5])$`)
	return re.MatchString(addr)
}

func ValidateDN(dn string) bool {
	re := regexp.MustCompile(`^(?:(?:CN|cn|OU|ou|DC|dc)\=[^,'"]+,)*(?:CN|cn|OU|ou|DC|dc)\=[^,'"]+$`)
	return re.MatchString(dn)
}
