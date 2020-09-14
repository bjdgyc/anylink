package dbdata

import (
	"encoding/json"
	"time"
)

const BucketUser = "user"

type User struct {
	Id        int
	Username  string
	Password  string
	OtpSecret string
	Group     []string
	// CreatedAt time.Time
	UpdatedAt time.Time
}

func GetUsers(lastKey string, prev bool) []User {
	res := getList(BucketUser, lastKey, prev)
	datas := make([]User, 0)
	for _, data := range res {
		d := User{}
		json.Unmarshal(data, &d)
		datas = append(datas, d)
	}
	return datas
}
