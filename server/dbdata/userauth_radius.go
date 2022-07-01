package dbdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type AuthRadius struct {
	Addr   string `json:"addr"`
	Secret string `json:"secret"`
}

func init() {
	authRegistry["radius"] = reflect.TypeOf(AuthRadius{})
}

func (auth AuthRadius) checkData(authData map[string]interface{}) error {
	authType := authData["type"].(string)
	bodyBytes, err := json.Marshal(authData[authType])
	if err != nil {
		return errors.New("Radius的密钥/服务器地址填写有误")
	}
	json.Unmarshal(bodyBytes, &auth)
	if !ValidateIpPort(auth.Addr) {
		return errors.New("Radius的服务器地址填写有误")
	}
	// freeradius官网最大8000字符, 这里限制200
	if len(auth.Secret) < 8 || len(auth.Secret) > 200 {
		return errors.New("Radius的密钥长度需在8～200个字符之间")
	}
	return nil
}

func (auth AuthRadius) checkUser(name, pwd string, g *Group) error {
	pl := len(pwd)
	if name == "" || pl < 1 {
		return fmt.Errorf("%s %s", name, "密码错误")
	}
	authType := g.Auth["type"].(string)
	if _, ok := g.Auth[authType]; !ok {
		return fmt.Errorf("%s %s", name, "Radius的radius值不存在")
	}
	bodyBytes, err := json.Marshal(g.Auth[authType])
	if err != nil {
		return fmt.Errorf("%s %s", name, "Radius Marshal出现错误")
	}
	err = json.Unmarshal(bodyBytes, &auth)
	if err != nil {
		return fmt.Errorf("%s %s", name, "Radius Unmarshal出现错误")
	}
	// radius认证时，设置超时3秒
	packet := radius.New(radius.CodeAccessRequest, []byte(auth.Secret))
	rfc2865.UserName_SetString(packet, name)
	rfc2865.UserPassword_SetString(packet, pwd)
	ctx, done := context.WithTimeout(context.Background(), 3*time.Second)
	defer done()
	response, err := radius.Exchange(ctx, packet, auth.Addr)
	if err != nil {
		return fmt.Errorf("%s %s", name, "Radius服务器连接异常, 请检测服务器和端口")
	}
	if response.Code != radius.CodeAccessAccept {
		return fmt.Errorf("%s %s", name, "Radius：用户名或密码错误")
	}
	return nil
}
