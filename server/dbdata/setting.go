package dbdata

import (
	"encoding/json"
	"reflect"
)

const (
	InstallName = "Install"
	InstallData = "OK"
)

type SettingSmtp struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	From       string `json:"from"`
	Encryption string `json:"encryption"`
}

type SettingOther struct {
	LinkAddr    string `json:"link_addr"`
	Banner      string `json:"banner"`
	AccountMail string `json:"account_mail"`
}

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

func SettingAdd(data interface{}) error {
	name := StructName(data)
	v, _ := json.Marshal(data)
	s := Setting{Name: name, Data: string(v)}
	err := Add(&s)
	return err
}

func SettingSet(data interface{}) error {
	name := StructName(data)
	v, _ := json.Marshal(data)
	s := Setting{Data: string(v)}
	err := Update("name", name, &s)
	return err
}

func SettingGet(data interface{}) error {
	name := StructName(data)
	s := Setting{Name: name}
	err := One("name", name, &s)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(s.Data), data)
	return err
}
