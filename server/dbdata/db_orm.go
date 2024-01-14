package dbdata

import (
	"errors"
	"reflect"

	"xorm.io/xorm"
)

const PageSize = 10

var ErrNotFound = errors.New("ErrNotFound")

func Add(data any) error {
	_, err := xdb.InsertOne(data)
	return err
}

func AddBatch(data any) error {
	_, err := xdb.Insert(data)
	return err
}

func Update(fieldName string, value any, data any) error {
	_, err := xdb.Where(fieldName+"=?", value).Update(data)
	return err
}

func Del(data any) error {
	_, err := xdb.Delete(data)
	return err
}

func extract(data any, fieldName string) any {
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
func Set(data any) error {
	id := extract(data, "Id")
	_, err := xdb.ID(id).AllCols().Update(data)
	return err
}

func One(fieldName string, value any, data any) error {
	has, err := xdb.Where(fieldName+"=?", value).Get(data)
	if err != nil {
		return err
	}
	if !has {
		return ErrNotFound
	}

	return nil
}

func CountAll(data any) int {
	n, _ := xdb.Count(data)
	return int(n)
}

func Find(data any, limit, page int) error {
	if limit == 0 {
		return xdb.Find(data)
	}

	start := (page - 1) * limit
	return xdb.Limit(limit, start).Find(data)
}

func FindWhere(data any, limit int, page int, where string, args ...any) error {
	if limit == 0 {
		return xdb.Where(where, args...).Find(data)
	}

	start := (page - 1) * limit
	return xdb.Where(where, args...).Limit(limit, start).Find(data)
}

func CountPrefix(fieldName string, prefix string, data any) int {
	n, _ := xdb.Where(fieldName+" like ?", prefix+"%").Count(data)
	return int(n)
}

func Prefix(fieldName string, prefix string, data any, limit, page int) error {
	where := xdb.Where(fieldName+" like ?", prefix+"%")
	if limit == 0 {
		return where.Find(data)
	}

	start := (page - 1) * limit
	return where.Limit(limit, start).Find(data)
}

func FindAndCount(session *xorm.Session, data any, limit, page int) (int64, error) {
	if limit == 0 {
		return session.FindAndCount(data)
	}
	start := (page - 1) * limit
	totalCount, err := session.Limit(limit, start).FindAndCount(data)
	return totalCount, err
}
