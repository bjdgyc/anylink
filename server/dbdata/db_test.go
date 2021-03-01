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
	assert := assert.New(t)
	preIpData()
	defer closeIpdata()

	u := User{Username: "a"}
	_ = Save(&u)

	assert.Equal(u.Id, 1)
}
