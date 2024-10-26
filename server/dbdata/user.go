package dbdata

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/xlzd/gotp"
	"golang.org/x/crypto/scrypt"
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

// 验证用户登录信息
func CheckUser(name, pwd, group string) error {
	// 获取登入的group数据
	groupData := &Group{}
	err := One("Name", group, groupData)
	if err != nil || groupData.Status != 1 {
		return fmt.Errorf("%s - %s", name, "用户组错误")
	}
	// 初始化Auth
	if len(groupData.Auth) == 0 {
		groupData.Auth["type"] = "local"
	}
	authType := groupData.Auth["type"].(string)
	// 本地认证方式
	if authType == "local" {
		return checkLocalUser(name, pwd, group)
	}
	// 其它认证方式, 支持自定义
	_, ok := authRegistry[authType]
	if !ok {
		return fmt.Errorf("%s %s", "未知的认证方式: ", authType)
	}
	auth := makeInstance(authType).(IUserAuth)
	return auth.checkUser(name, pwd, groupData)
}

// 验证本地用户登录信息
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
		switch v.Status {
		case 0:
			return fmt.Errorf("%s %s", name, "用户不存在或用户已停用")
		case 2:
			return fmt.Errorf("%s %s", name, "用户已过期")
		}
	}
	// 判断用户组信息
	if !utils.InArrStr(v.Groups, group) {
		return fmt.Errorf("%s %s", name, "用户组错误")
	}
	// 判断otp信息
	// pinCode := pwd
	// if !v.DisableOtp {
	// 	pinCode = pwd[:pl-6]
	// 	otp := pwd[pl-6:]
	// 	if !CheckOtp(name, otp, v.OtpSecret) {
	// 		return fmt.Errorf("%s %s", name, "动态码错误")
	// 	}
	// }
	// 判断用户密码
	if !VerifyPassword(pwd, v.PinCode) {
		return fmt.Errorf("%s %s", name, "密码错误")
	}

	return nil
}

// 用户过期时间到达后，更新用户状态，并返回一个状态为过期的用户切片
func CheckUserlimittime() (limitUser []interface{}) {
	if _, err := xdb.Where("limittime <= ?", time.Now()).And("status = ?", 1).Update(&User{Status: 2}); err != nil {
		return
	}
	user := make(map[int64]User)
	if err := xdb.Where("status != ?", 1).Find(user); err != nil {
		return
	}
	for _, v := range user {
		limitUser = append(limitUser, v.Username)
	}
	return
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
func CheckOtp(name, otp, secret string) bool {
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
	verify := totp.Verify(otp, unix)

	return verify
}

// 插入数据库前加密密码
func (u *User) BeforeInsert() error {
	hashedPassword, err := ScryptPassword(u.PinCode)
	if err != nil {
		base.Error(err)
		return err
	}
	u.PinCode = hashedPassword
	return nil
}

// 更新数据库前加密密码
func (u *User) BeforeUpdate() error {
	if len(u.PinCode) != 57 {
		hashedPassword, err := ScryptPassword(u.PinCode)
		if err != nil {
			base.Error(err)
			return err
		}
		u.PinCode = hashedPassword
	}
	return nil
}

// 加密密码
func ScryptPassword(passwd string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hashPasswd, err := scrypt.Key([]byte(passwd), salt, 1<<16, 8, 1, 32)
	if err != nil {
		return "", err
	}

	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedHash := base64.StdEncoding.EncodeToString(hashPasswd)

	return encodedSalt + "&" + encodedHash, nil
}

// 验证密码
func VerifyPassword(password, hashPassword string) bool {
	// 老用户使用明文验证
	if len(hashPassword) != 57 {
		return password == hashPassword
	}

	// 分割盐值和哈希值
	encodepwds := strings.SplitN(hashPassword, "&", 2)
	if len(encodepwds) != 2 {
		return false
	}

	// 解码盐值
	salt, err := base64.StdEncoding.DecodeString(encodepwds[0])
	if err != nil {
		return false
	}

	// 计算新的哈希值
	newHash, err := scrypt.Key([]byte(password), salt, 1<<16, 8, 1, 32)
	if err != nil {
		return false
	}

	return base64.StdEncoding.EncodeToString(newHash) == encodepwds[1]
}
