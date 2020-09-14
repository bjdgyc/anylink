package dbdata

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/bjdgyc/anylink/common"
	bolt "go.etcd.io/bbolt"
)

const pageSize = 10

var (
	db       *bolt.DB
	ErrNoKey = errors.New("db no this key")
)

func initDb() {
	var err error
	db, err = bolt.Open(common.ServerCfg.DbFile, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 创建bucket
	err = db.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists([]byte(BucketUser))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(BucketGroup))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(BucketMacIp))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func NextId(bucket string) int {
	var i int
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		id, err := b.NextSequence()
		i = int(id)
		// discard error
		return err
	})
	return i
}

func GetCount(bucket string) int {
	count := 0
	db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		s := bkt.Stats()
		// fmt.Printf("%+v \n", s)
		count = s.KeyN
		return nil
	})
	return count
}

func Set(bucket, key string, v interface{}) error {
	return db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return bkt.Put([]byte(key), b)
	})
}

func Del(bucket, key string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		return bkt.Delete([]byte(key))
	})
}

func Get(bucket, key string, v interface{}) error {
	return db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		b := bkt.Get([]byte(key))
		if b == nil {
			return ErrNoKey
		}
		return json.Unmarshal(b, v)
	})
}

// 分页获取
func getList(bucket, lastKey string, prev bool) [][]byte {
	res := make([][]byte, 0)
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucket)).Cursor()
		size := pageSize
		k, b := c.Seek([]byte(lastKey))

		if prev {
			for i := 0; i < size; i++ {
				k, b = c.Prev()
				if k == nil {
					break
				}
				res = append(res, b)
			}
			return nil
		}

		// next
		if string(k) != lastKey {
			// 不相同，说明找出其他的
			size -= 1
			res = append(res, b)
		}
		for i := 0; i < size; i++ {
			k, b = c.Next()
			if k == nil {
				break
			}
			res = append(res, b)
		}
		return nil
	})
	return res
}
