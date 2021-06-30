package dbdata

const PageSize = 10

func Save(data interface{}) error {
	_, err := x.Insert(data)
	if err != nil {
		return err
	}
	return nil
}

func Del(data interface{}) error {
	_, err := x.Delete(data)
	if err != nil {
		return err
	}
	return nil
}

func Set(fieldName string, value interface{}, data interface{}) error {
	_, err := x.AllCols().Where(fieldName+"=?", value).Update(data)
	if err != nil {
		return err
	}
	return nil
}

func Get(fieldName string, value interface{}, data interface{}) error {
	_, err := x.Where(fieldName+"=?", value).Get(data)
	if err != nil {
		return err
	}
	return nil
}

func CountAll(data interface{}) int {
	n, _ := x.Count(data)
	return int(n)
}

func One(fieldName string, value interface{}, to interface{}) (bool, error) {
	// _, err := x.Where(fieldName+"=?", value).Get(to)
	// if err != nil {
	// 	return err
	// }
	return x.Where(fieldName+"=?", value).Get(to)
}

func All(to interface{}, limit, page int) (int64, error) {

	return x.Limit(limit, page-1).FindAndCount(to)
}

func Prefix(fieldName string, prefix string, to interface{}, limit, page int) error {

	err := x.Where(fieldName+" LIKE ?", "%"+prefix+"%").Limit(limit, page-1).Find(to)
	if err != nil {
		return err
	}
	return nil
}
