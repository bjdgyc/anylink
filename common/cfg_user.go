package common

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/pelletier/go-toml"
)

var (
	users       = map[string]User{}
	limitClient = map[string]int{"_all": 0}
	limitMux    = sync.Mutex{}
)

type User struct {
	Group     string `toml:"group"`
	Username  string `toml:"-"`
	Password  string `toml:"password"`
	OtpSecret string `toml:"otp_secret"`
}

func CheckUser(name, pwd, group string) bool {
	user, ok := users[name]
	if !ok {
		return false
	}
	pwdHash := hashPass(pwd)
	if user.Password == pwdHash {
		return true
	}
	return false
}

func hashPass(pwd string) string {
	sum := sha1.Sum([]byte(pwd))
	return fmt.Sprintf("%x", sum)
}

func LimitClient(name string, close bool) bool {
	limitMux.Lock()
	defer limitMux.Unlock()
	// defer fmt.Println(limitClient)

	_all := limitClient["_all"]
	c, ok := limitClient[name]
	if !ok { // 不存在用户
		limitClient[name] = 0
	}

	if close {
		limitClient[name] = c - 1
		limitClient["_all"] = _all - 1
		return true
	}

	// 全局判断
	if _all >= ServerCfg.MaxClient {
		return false
	}

	// 超出同一个用户限制
	if c >= ServerCfg.MaxUserClient {
		return false
	}

	limitClient[name] = c + 1
	limitClient["_all"] = _all + 1
	return true
}

func loadUser() {
	b, err := ioutil.ReadFile(ServerCfg.UserFile)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(b, &users)
	if err != nil {
		panic(err)
	}

	// 添加用户名
	for k, v := range users {
		v.Username = k
		users[k] = v
	}

	fmt.Println("users:", users)
}
