package dbdata

import (
	"encoding/json"
	"reflect"

	"xorm.io/xorm"
)

type SettingInstall struct {
	Installed bool `json:"installed"`
}

type SettingSmtp struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	From       string `json:"from"`
	Encryption string `json:"encryption"`
}

type SettingAuditLog struct {
	AuditInterval int    `json:"audit_interval"`
	LifeDay       int    `json:"life_day"`
	ClearTime     string `json:"clear_time"`
}

type SettingOther struct {
	LinkAddr    string `json:"link_addr"`
	Banner      string `json:"banner"`
	Homeindex   string `json:"homeindex"`
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

func SettingSessAdd(sess *xorm.Session, data interface{}) error {
	name := StructName(data)
	v, _ := json.Marshal(data)
	s := &Setting{Name: name, Data: v}
	_, err := sess.InsertOne(s)
	return err
}

func SettingSet(data interface{}) error {
	name := StructName(data)
	v, _ := json.Marshal(data)
	s := &Setting{Data: v}
	err := Update("name", name, s)
	return err
}

func SettingGet(data interface{}) error {
	name := StructName(data)
	s := &Setting{}
	err := One("name", name, s)
	if err != nil {
		return err
	}
	err = json.Unmarshal(s.Data, data)
	return err
}

func SettingGetAuditLog() (SettingAuditLog, error) {
	data := SettingAuditLog{}
	err := SettingGet(&data)
	if err == nil {
		return data, err
	}
	if !CheckErrNotFound(err) {
		return data, err
	}
	sess := xdb.NewSession()
	defer sess.Close()
	auditLog := SettingGetAuditLogDefault()
	err = SettingSessAdd(sess, auditLog)
	if err != nil {
		return data, err
	}
	return auditLog, nil
}

func SettingGetAuditLogDefault() SettingAuditLog {
	auditLog := SettingAuditLog{
		LifeDay:   0,
		ClearTime: "05:00",
	}
	return auditLog
}
