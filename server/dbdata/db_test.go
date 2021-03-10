package dbdata

import (
	"os"
	"path"
	"testing"

	"github.com/bjdgyc/anylink/base"
	"github.com/stretchr/testify/assert"
)

func preIpData() {
	tmpDb := path.Join(os.TempDir(), "anylink_test.db")
	base.Cfg.DbFile = tmpDb
	initDb()
}

func closeIpdata() {
	sdb.Close()
	tmpDb := path.Join(os.TempDir(), "anylink_test.db")
	os.Remove(tmpDb)
}

func TestDb(t *testing.T) {
	ast := assert.New(t)
	preIpData()
	defer closeIpdata()

	u := User{Username: "a"}
	err := Save(&u)
	ast.Nil(err)

	ast.Equal(u.Id, 1)
}
