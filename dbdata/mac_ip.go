package dbdata

import (
	"encoding/json"
	"net"
	"time"

	bolt "go.etcd.io/bbolt"
)

const BucketMacIp = "macIp"

type MacIp struct {
	IsActive  bool // db存储没有使用
	Ip        net.IP
	MacAddr   string
	LastLogin time.Time
}

func GetAllMacIp() []MacIp {
	datas := make([]MacIp, 0)
	db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(BucketMacIp))
		bkt.ForEach(func(k, v []byte) error {
			d := MacIp{}
			json.Unmarshal(v, &d)
			datas = append(datas, d)
			return nil
		})
		return nil
	})

	return datas
}
