package dbdata

import "reflect"

var authRegistry = make(map[string]reflect.Type)

type IUserAuth interface {
	checkData(authData map[string]interface{}) error
	checkUser(name string, pwd string, authData map[string]interface{}) error
}

func makeInstance(name string) interface{} {
	v := reflect.New(authRegistry[name]).Elem()
	return v.Interface()
}
