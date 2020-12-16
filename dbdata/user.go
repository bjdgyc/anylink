package dbdata

import (
	"errors"
	"time"

	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/xlzd/gotp"
)

type User struct {
	Id       int    `json:"id" storm:"id,increment"`
	Username string `json:"username" storm:"unique"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	// Password  string    `json:"password"`
	PinCode   string    `json:"pin_code"`
	OtpSecret string    `json:"otp_secret"`
	Groups    []string  `json:"groups"`
	Status    int8      `json:"status"` // 1正常
	SendEmail bool      `json:"send_email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 验证用户登陆信息
func CheckUser(name, pwd, group string) error {
	// return nil

	pl := len(pwd)
	if name == "" || pl < 6 {
		return errors.New("密码错误")
	}
	v := &User{}
	err := One("Username", name, v)
	if err != nil || v.Status != 1 {
		return errors.New("用户名错误")
	}
	pass := pwd[:pl-6]
	// if !utils.PasswordVerify(pass, v.Password) {
	if pass != v.PinCode {
		return errors.New("密码错误")
	}
	otp := pwd[pl-6:]
	totp := gotp.NewDefaultTOTP(v.OtpSecret)
	unix := time.Now().Unix()
	verify := totp.Verify(otp, int(unix))
	if !verify {
		return errors.New("动态码错误")
	}

	// 判断用户组信息
	if !utils.InArrStr(v.Groups, group) {
		return errors.New("用户组错误")
	}
	groupData := &Group{}
	err = One("Name", group, groupData)
	if err != nil || groupData.Status != 1 {
		return errors.New("用户组错误")
	}
	return nil
}

func SetUser(v *User) error {
	var err error
	if v.Username == "" || len(v.Groups) == 0 {
		return errors.New("用户名或组错误")
	}

	planPass := v.PinCode
	// 自动生成密码
	if len(planPass) < 6 {
		planPass = utils.RandomNum(8)
	}
	v.PinCode = planPass

	if v.OtpSecret == "" {
		v.OtpSecret = gotp.RandomSecret(24)
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
	err = Save(v)

	return err
}
