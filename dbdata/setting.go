package dbdata

import (
	"reflect"
)

const (
	SettingBucket = "SettingBucket"
	Installed     = "Installed"
)

func StructName(data interface{}) string {
	ref := reflect.ValueOf(data)
	s := &ref
	if s.Kind() == reflect.Ptr {
		e := s.Elem()
		s = &e
	}
	name := s.Type().Name()
	return name
}

func SettingSet(data interface{}) error {
	key := StructName(data)
	err := Set(SettingBucket, key, data)
	return err
}

func SettingGet(data interface{}) error {
	key := StructName(data)
	err := Get(SettingBucket, key, data)
	return err
}

type SettingSmtp struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

type SettingOther struct {
	Banner      string `json:"banner"`
	AccountMail string `json:"account_mail"`
}
