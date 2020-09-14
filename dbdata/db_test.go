package dbdata

import (
	"net"
	"os"
	"path"
	"testing"

	"github.com/bjdgyc/anylink/common"
	"github.com/stretchr/testify/assert"
)

func preIpData() {
	tmpDb := path.Join(os.TempDir(), "anylink_test.db")
	common.ServerCfg.DbFile = tmpDb
	initDb()
}

func closeIpdata() {
	db.Close()
	tmpDb := path.Join(os.TempDir(), "anylink_test.db")
	os.Remove(tmpDb)
}

func TestDb(t *testing.T) {
	assert := assert.New(t)
	preIpData()
	defer closeIpdata()

	Set(BucketUser, "a", User{Username: "a"})
	Set(BucketUser, "b", User{Username: "b"})
	Set(BucketUser, "c", User{Username: "c"})
	Set(BucketUser, "d", User{Username: "d"})
	Set(BucketUser, "e", User{Username: "e"})
	Set(BucketUser, "f", User{Username: "f"})
	Set(BucketUser, "g", User{Username: "g"})

	c := GetCount(BucketUser)
	assert.Equal(c, 7)
	Del(BucketUser, "g")
	c = GetCount(BucketUser)
	assert.Equal(c, 6)

	// 分页查询
	us := GetUsers("d", false)
	assert.Equal(us[0].Username, "e")
	assert.Equal(us[1].Username, "f")
	us = GetUsers("d", true)
	assert.Equal(us[0].Username, "c")
	assert.Equal(us[1].Username, "b")
	assert.Equal(us[2].Username, "a")

	mac1 := MacIp{Ip: net.ParseIP("192.168.3.11"), MacAddr: "mac1"}
	mac2 := MacIp{Ip: net.ParseIP("192.168.3.12"), MacAddr: "mac2"}
	Set(BucketMacIp, "mac1", mac1)
	Set(BucketMacIp, "mac2", mac2)

	mp := GetAllMacIp()
	assert.Equal(mp[0].MacAddr, "mac1")
	assert.Equal(mp[1].MacAddr, "mac2")

	os.Exit(0)
}
