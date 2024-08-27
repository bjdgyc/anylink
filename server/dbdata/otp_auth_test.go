package dbdata

import (
	"fmt"
	. "github.com/go-ldap/ldap/v3"
	"github.com/spf13/viper"
	"strconv"
	"testing"
)

var attributes = []string{
	"cn",
	"sAMAccountName",
	"displayName",
}

func TestUserOtpAuth(t *testing.T) {

	v := viper.New()
	v.SetConfigFile("../conf/server.toml")
	if err := v.ReadInConfig(); err != nil {
		panic("config file err:" + err.Error())
	}

	user, pwd := v.Get("ldap_user").(string), v.Get("ldap_pass").(string)
	addr, baseDN := v.Get("ldap_server").(string), v.Get("ldap_base_dn").(string)

	l, err := DialURL(fmt.Sprintf("ldap://%s", addr))
	if err != nil {
		t.Fatal(err)
	}

	defer l.Close()

	err = l.Bind(user, pwd)
	if err != nil {
		t.Fatal(err)
	}

	user = "test"
	searchRequest := NewSearchRequest(
		baseDN,
		ScopeWholeSubtree, DerefAlways, 0, 0, false,
		fmt.Sprintf("(&(objectClass=person)(sAMAccountName=%s))", user),
		attributes,
		nil)

	sr, err := l.Search(searchRequest)
	if err != nil {
		t.Fatal(err)
	}

	userDN := sr.Entries[0].DN
	fmt.Println("userDN: ", userDN)

	pwd = "tests1sDSs$872322"
	pl := len(pwd)
	pinCode := pwd[:pl-6]
	otp := pwd[pl-6:]

	err = l.Bind(userDN, pinCode)
	if err != nil {
		t.Fatalf("LDAP 登入失败，请检查登入的账号 [%s] 或密码 [%v], err=[%v]", userDN, pinCode, err.Error())
	} else {
		// check user otp
		ot, err := strconv.Atoi(otp)
		otpAuthRes, err := ValidateUserOtp(user, ot)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("otpAuthRes: ", otpAuthRes)
	}
}
