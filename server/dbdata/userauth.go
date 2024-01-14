package dbdata

import (
	"reflect"
	"regexp"
)

var authRegistry = make(map[string]reflect.Type)

type IUserAuth interface {
	checkData(authData map[string]any) error
	checkUser(name, pwd string, g *Group) error
}

func makeInstance(name string) any {
	v := reflect.New(authRegistry[name]).Elem()
	return v.Interface()
}

func ValidateIpPort(addr string) bool {
	RegExp := regexp.MustCompile(`^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\:([0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-5]{2}[0-3][0-5])$$`)
	return RegExp.MatchString(addr)
}
