package dbdata

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/xlzd/gotp"
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
func CheckUser(name, pwd string, group *string) error {
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
	// 当只有一个可用的用户组时，客户端不显示用户组选项，默认使用该用户唯一可用的组
	if *group == "" {
		Groups := GetAvailableGroups()
		for _, g := range v.Groups {
			if utils.InArrStr(Groups, g) {
				*group = g
				break
			}
		}
	} else if !utils.InArrStr(v.Groups, *group) {
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
