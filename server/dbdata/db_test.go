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
	base.Cfg.DbType = "sqlite3"
	base.Cfg.DbSource = tmpDb
	initDb()
}

func closeIpdata() {
	xdb.Close()
	tmpDb := path.Join(os.TempDir(), "anylink_test.db")
	os.Remove(tmpDb)
}

func TestDb(t *testing.T) {
	ast := assert.New(t)
	preIpData()
	defer closeIpdata()

	u := User{Username: "a"}
	err := Add(&u)
	ast.Nil(err)

	ast.Equal(u.Id, 1)
}
