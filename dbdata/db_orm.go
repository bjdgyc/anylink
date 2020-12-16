package dbdata

import "github.com/asdine/storm/v3/index"

const PageSize = 10

func Save(data interface{}) error {
	return sdb.Save(data)
}

func Update(data interface{}) error {
	return sdb.Update(data)
}

func UpdateField(data interface{}, fieldName string, value interface{}) error {
	return sdb.UpdateField(data, fieldName, value)
}

func Del(data interface{}) error {
	return sdb.DeleteStruct(data)
}

func Set(bucket, key string, data interface{}) error {
	return sdb.Set(bucket, key, data)
}

func Get(bucket, key string, data interface{}) error {
	return sdb.Get(bucket, key, data)
}

func CountAll(data interface{}) int {
	n, _ := sdb.Count(data)
	return n
}

func One(fieldName string, value interface{}, to interface{}) error {
	return sdb.One(fieldName, value, to)
}

func Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error {
	return sdb.Find(fieldName, value, to, options...)
}

func All(to interface{}, limit, page int) error {
	opt := getOpt(limit, page)
	return sdb.All(to, opt)
}

func Prefix(fieldName string, prefix string, to interface{}, limit, page int) error {
	opt := getOpt(limit, page)
	return sdb.Prefix(fieldName, prefix, to, opt)
}

func getOpt(limit, page int) func(*index.Options) {
	skip := (page - 1) * limit
	opt := func(opt *index.Options) {
		opt.Reverse = true
		if limit > 0 {
			opt.Limit = limit
		}
		if skip > 0 {
			opt.Skip = skip
		}
	}
	return opt
}
