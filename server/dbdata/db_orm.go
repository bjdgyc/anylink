package dbdata

import (
	"errors"
	"reflect"
)

const PageSize = 10

var ErrNotFound = errors.New("ErrNotFound")

func Add(data interface{}) error {
	_, err := xdb.InsertOne(data)
	return err
}

func Update(fieldName string, value interface{}, data interface{}) error {
	_, err := xdb.Where(fieldName+"=?", value).Update(data)
	return err
}

func Del(data interface{}) error {
	_, err := xdb.Delete(data)
	return err
}

func extract(data interface{}, fieldName string) interface{} {
	ref := reflect.ValueOf(data)
	r := &ref
	if r.Kind() == reflect.Ptr {
		e := r.Elem()
		r = &e
	}
	field := r.FieldByName(fieldName).Interface()
	return field
}

// 更新全部字段
func Set(data interface{}) error {
	id := extract(data, "Id")
	_, err := xdb.ID(id).AllCols().Update(data)
	return err
}

func One(fieldName string, value interface{}, data interface{}) error {
	has, err := xdb.Where(fieldName+"=?", value).Get(data)
	if err != nil {
		return err
	}
	if !has {
		return ErrNotFound
	}

	return nil
}

func CountAll(data interface{}) int {
	n, _ := xdb.Count(data)
	return int(n)
}

func Find(data interface{}, limit, page int) error {
	if limit == 0 {
		return xdb.Find(data)
	}

	start := (page - 1) * limit
	return xdb.Limit(limit, start).Find(data)
}

func CountPrefix(fieldName string, prefix string, data interface{}) int {
	n, _ := xdb.Where(fieldName + " like '" + prefix + "%' ").Count(data)
	return int(n)
}

func Prefix(fieldName string, prefix string, data interface{}, limit, page int) error {
	where := xdb.Where(fieldName + " like '" + prefix + "%' ")
	if limit == 0 {
		return where.Find(data)
	}

	start := (page - 1) * limit
	return where.Limit(limit, start).Find(data)
}
