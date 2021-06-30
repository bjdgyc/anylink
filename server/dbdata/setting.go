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

func StructKey(data interface{}) map[string]interface{} {
	ref := reflect.ValueOf(data)
	s := &ref
	if s.Kind() == reflect.Ptr {
		e := s.Elem()
		s = &e
	}
	d1 := make(map[string]interface{}, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		//fmt.Printf("%d. %s %s = %v \n", i, s.Type().Field(i).Name, field.Type(), field.Interface())
		d1[s.Type().Field(i).Name] = field.Interface()
	}
	return d1
}

func SettingSet(data interface{}) error {

	err := Set("id", 1, data)
	return err

}

func SettingGet(data interface{}) error {
	//key := StructName(data)
	err := Get("id", 1, data)
	return err
}

func CheckErrNotFound(err error) bool {
	// if fmt.Sprint(err) == "0" {
	// 	return false
	// }
	return true
}
