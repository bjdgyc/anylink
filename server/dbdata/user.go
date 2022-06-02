package dbdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/xlzd/gotp"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// type User struct {
// 	Id       int    `json:"id"  xorm:"pk autoincr not null"`
// 	Username string `json:"username" storm:"not null unique"`
// 	Nickname string `json:"nickname"`
// 	Email    string `json:"email"`
// 	// Password  string    `json:"password"`
// 	PinCode    string    `json:"pin_code"`
// 	OtpSecret  string    `json:"otp_secret"`
// 	DisableOtp bool      `json:"disable_otp"` // 禁用otp
// 	Groups     []string  `json:"groups"`
// 	Status     int8      `json:"status"` // 1正常
// 	SendEmail  bool      `json:"send_email"`
// 	CreatedAt  time.Time `json:"created_at"`
// 	UpdatedAt  time.Time `json:"updated_at"`
// }

func SetUser(v *User) error {
	var err error
	if v.Username == "" || len(v.Groups) == 0 {
		return errors.New("用户名或组错误")
	}

	planPass := v.PinCode
	// 自动生成密码
	if len(planPass) < 6 {
		planPass = utils.RandomRunes(8)
	}
	v.PinCode = planPass

	if v.OtpSecret == "" {
		v.OtpSecret = gotp.RandomSecret(32)
	}

	// 判断组是否有效
	ng := []string{}
	groups := GetGroupNames()
	for _, g := range v.Groups {
		if utils.InArrStr(groups, g) {
			ng = append(ng, g)
		}
	}
	if len(ng) == 0 {
		return errors.New("用户名或组错误")
	}
	v.Groups = ng

	v.UpdatedAt = time.Now()
	if v.Id > 0 {
		err = Set(v)
	} else {
		err = Add(v)
	}

	return err
}

// 验证用户登陆信息
func CheckUser(name, pwd, group string) error {
	// 获取登入的group数据
	groupData := &Group{}
	err := One("Name", group, groupData)
	if err != nil {
		return fmt.Errorf("%s %s", name, "No用户组")
	}
	if len(groupData.Auth) == 0 {
		groupData.Auth["type"] = "local"
	}
	base.Debug(name + " auth type: " + fmt.Sprintf("%s", groupData.Auth["type"]))
	switch groupData.Auth["type"] {
	case "local":
		return checkLocalUser(name, pwd, group)
	case "radius":
		radisConf := AuthRadius{}
		bodyBytes, err := json.Marshal(groupData.Auth["radius"])
		if err != nil {
			fmt.Errorf("%s %s", name, "Radius出现Marshal错误")
		}
		err = json.Unmarshal(bodyBytes, &radisConf)
		if err != nil {
			fmt.Errorf("%s %s", name, "Radius出现Unmarshal错误")
		}
		return checkRadiusUser(name, pwd, radisConf)
	default:
		return fmt.Errorf("%s %s", name, "无效的认证类型")
	}
	return nil
}

// 验证本地用户登陆信息
func checkLocalUser(name, pwd, group string) error {
	// TODO 严重问题
	// return nil

	pl := len(pwd)
	if name == "" || pl < 6 {
		return fmt.Errorf("%s %s", name, "密码错误")
	}
	v := &User{}
	err := One("Username", name, v)
	if err != nil || v.Status != 1 {
		return fmt.Errorf("%s %s", name, "用户名错误")
	}
	// 判断用户组信息
	if !utils.InArrStr(v.Groups, group) {
		return fmt.Errorf("%s %s", name, "用户组错误")
	}
	groupData := &Group{}
	err = One("Name", group, groupData)
	if err != nil || groupData.Status != 1 {
		return fmt.Errorf("%s - %s", name, "用户组错误")
	}

	// 判断otp信息
	pinCode := pwd
	if !v.DisableOtp {
		pinCode = pwd[:pl-6]
		otp := pwd[pl-6:]
		if !checkOtp(name, otp, v.OtpSecret) {
			return fmt.Errorf("%s %s", name, "动态码错误")
		}
	}

	// 判断用户密码
	if pinCode != v.PinCode {
		return fmt.Errorf("%s %s", name, "密码错误")
	}

	return nil
}

func checkRadiusUser(name string, pwd string, raduisConf AuthRadius) error {
	packet := radius.New(radius.CodeAccessRequest, []byte(raduisConf.Secret))
	rfc2865.UserName_SetString(packet, name)
	rfc2865.UserPassword_SetString(packet, pwd)
	ctx, done := context.WithTimeout(context.Background(), 3*time.Second)
	defer done()
	response, err := radius.Exchange(ctx, packet, raduisConf.Addr)
	if err != nil {
		return fmt.Errorf("%s %s", name, "Radius服务器连接异常, 请检测服务器和端口")
	}
	if response.Code != radius.CodeAccessAccept {
		return fmt.Errorf("%s %s", name, "Radius：用户名或密码错误")
	}
	return nil
}

var (
	userOtpMux = sync.Mutex{}
	userOtp    = map[string]time.Time{}
)

func init() {
	go func() {
		expire := time.Second * 60

		for range time.Tick(time.Second * 10) {
			tnow := time.Now()
			userOtpMux.Lock()
			for k, v := range userOtp {
				if tnow.After(v.Add(expire)) {
					delete(userOtp, k)
				}
			}
			userOtpMux.Unlock()
		}
	}()
}

// 判断令牌信息
func checkOtp(name, otp, secret string) bool {
	key := fmt.Sprintf("%s:%s", name, otp)

	userOtpMux.Lock()
	defer userOtpMux.Unlock()

	// 令牌只能使用一次
	if _, ok := userOtp[key]; ok {
		// 已经存在
		return false
	}
	userOtp[key] = time.Now()

	totp := gotp.NewDefaultTOTP(secret)
	unix := time.Now().Unix()
	verify := totp.Verify(otp, int(unix))

	return verify
}
