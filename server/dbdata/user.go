package dbdata

import (
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xlzd/gotp"
)

func SetUser(v *User) error {
	var err error
	if v.Username == "" || len(v.Groups) == 0 {
		return errors.New("用户名或组错误")
	}

	planPass := v.PinCode
	// 自动生成密码
	if len(planPass) < 6 {
		planPass = RandomRunes(8)
	}
	v.PinCode = planPass

	// 判断组是否有效
	ng := []string{}
	groups := GetGroupNames()
	for _, g := range v.Groups {
		if InArrStr(groups, g) {
			ng = append(ng, g)
		}
	}
	if len(ng) == 0 {
		return errors.New("用户名或组错误")
	}
	v.Groups = ng

	if v.Id == 0 {
		v.UpdatedAt = time.Now()
		v.CreatedAt = time.Now()

		v.OtpSecret = gotp.RandomSecret(32)

		err = Save(v)
	} else {

		if v.OtpSecret == "" {
			v.OtpSecret = gotp.RandomSecret(32)
		}
		v.UpdatedAt = time.Now()
		err = Set("id", v.Id, v)

	}

	return err
}

// 验证用户登陆信息
func CheckUser(name, pwd, group, macAddress string) (username, groupstr, macaddress string, err error) {
	// TODO 严重问题
	// return nil
	d1, ok := autocache.Get(macAddress)
	if ok {

		autodata, ok := d1.(map[string]string)
		autocache.Delete(macAddress)
		if !ok {
			return autodata["name"], autodata["group"], autodata["macAddress"], fmt.Errorf("%s %s", name, "认证数据错误")
		}
		if !checkOtp(autodata["name"], pwd, autodata["otp"]) {

			return autodata["name"], autodata["group"], autodata["macAddress"], fmt.Errorf("%s %s", name, "动态码错误")
		}

		return autodata["name"], autodata["group"], autodata["macAddress"], nil

	} else {
		v := &User{}
		ok, err := One("Username", name, v)
		if err != nil || !ok {
			return name, group, macAddress, fmt.Errorf("%s %s", name, "用户名错误")
		}
		if !v.DisableOtp {

			pl := len(pwd)
			if name == "" || pl < 6 {
				return name, group, macAddress, fmt.Errorf("%s %s", name, "密码错误")
			}

			// 判断用户组信息
			if !InArrStr(v.Groups, group) {
				fmt.Printf("%+v\n%s\n", *v, group)
				return name, group, macAddress, fmt.Errorf("%s %s", name, "用户组错误")
			}
			groupData := &Group{}
			ok, err := One("Name", group, groupData)

			if err != nil || !ok {
				fmt.Printf("%+v", *groupData)
				return name, group, macAddress, fmt.Errorf("%s - %s", name, "用户组错误")
			}

			if pwd != v.PinCode {
				return name, group, macAddress, fmt.Errorf("%s %s", name, "密码错误")
			}
			t1 := make(map[string]string, 3)
			t1["name"] = name
			t1["password"] = pwd
			t1["group"] = group
			t1["otp"] = v.OtpSecret
			autocache.Set(macAddress, t1, cache.DefaultExpiration)
			return name, group, macAddress, fmt.Errorf("%s", "otpcheck")
		} else {
			// 判断用户密码

			pl := len(pwd)
			if name == "" || pl < 6 {
				return name, group, macAddress, fmt.Errorf("%s %s", name, "密码错误")
			}

			// 判断用户组信息
			if !InArrStr(v.Groups, group) {
				fmt.Printf("%+v\n%s\n", *v, group)
				return name, group, macAddress, fmt.Errorf("%s %s", name, "用户组错误")
			}
			groupData := &Group{}
			ok, err := One("Name", group, groupData)

			if err != nil || !ok {
				fmt.Printf("%+v", *groupData)
				return name, group, macAddress, fmt.Errorf("%s - %s", name, "用户组错误")
			}

			if pwd != v.PinCode {
				return name, group, macAddress, fmt.Errorf("%s %s", name, "密码错误")
			}
			return name, group, macAddress, nil
		}

	}

	// 判断otp信息
	// pinCode := pwd
	// if !v.DisableOtp {
	// 	pinCode = pwd[:pl-6]
	// 	otp := pwd[pl-6:]
	// 	if !checkOtp(name, otp, v.OtpSecret) {
	// 		return fmt.Errorf("%s %s", name, "动态码错误")
	// 	}
	// }

}

var (
	userOtp = New()
)

func init() {
	go func() {
		expire := time.Second * 60

		for range time.Tick(time.Second * 10) {
			tnow := time.Now()

			for v := range userOtp.IterBuffered() {
				if tnow.After(v.Val.(time.Time).Add(expire)) {
					userOtp.Remove(v.Key)
				}
			}

		}
	}()
}

// 判断令牌信息
func checkOtp(name, otp, secret string) bool {
	key := fmt.Sprintf("%s:%s", name, otp)

	// 令牌只能使用一次
	if _, ok := userOtp.Get(key); ok {
		// 已经存在
		return false
	}
	userOtp.Set(key, time.Now())

	totp := gotp.NewDefaultTOTP(secret)
	unix := time.Now().Unix()
	verify := totp.Verify(otp, int(unix))

	return verify
}
