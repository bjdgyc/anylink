package sessdata

import (
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/stretchr/testify/assert"
)

func preData(tmpDir string) {
	base.Test()
	tmpDb := path.Join(tmpDir, "test.db")
	base.Cfg.DbType = "sqlite3"
	base.Cfg.DbSource = tmpDb
	base.Cfg.Ipv4CIDR = "192.168.3.0/24"
	base.Cfg.Ipv4Gateway = "192.168.3.1"
	base.Cfg.Ipv4Start = "192.168.3.2"
	base.Cfg.Ipv4End = "192.168.3.199"
	base.Cfg.MaxClient = 100
	base.Cfg.MaxUserClient = 3
	base.Cfg.IpLease = 10

	dbdata.Start()
	group := dbdata.Group{
		Name:      "group1",
		Bandwidth: 1000,
	}
	_ = dbdata.Add(&group)
	initIpPool()
}

func cleardata(tmpDir string) {
	_ = dbdata.Stop()
	tmpDb := path.Join(tmpDir, "test.db")
	os.Remove(tmpDb)
}

func TestIpPool(t *testing.T) {
	assert := assert.New(t)
	tmp := t.TempDir()
	preData(tmp)
	defer cleardata(tmp)

	var ip net.IP

	for i := 2; i <= 100; i++ {
		_ = AcquireIp(getTestUser(i), getTestMacAddr(i), true)
	}

	ip = AcquireIp(getTestUser(101), getTestMacAddr(101), true)
	assert.True(net.IPv4(192, 168, 3, 101).Equal(ip))
	for i := 102; i <= 199; i++ {
		ip = AcquireIp(getTestUser(i), getTestMacAddr(i), true)
	}
	assert.True(net.IPv4(192, 168, 3, 199).Equal(ip))
	ip = AcquireIp(getTestUser(200), getTestMacAddr(200), true)
	assert.Nil(ip)

	// 回收188
	ReleaseIp(net.IPv4(192, 168, 3, 188), getTestMacAddr(188))
	time.Sleep(time.Second * 15)

	// 从头循环获取可用ip
	user_new := getTestUser(210)
	mac_new := getTestMacAddr(210)
	ip = AcquireIp(user_new, mac_new, true)
	t.Log("mac_new", ip)
	assert.True(net.IPv4(192, 168, 3, 188).Equal(ip))
}

func getTestUser(i int) string {
	return fmt.Sprintf("user-%d", i)
}

func getTestMacAddr(i int) string {
	// 前缀mac
	macAddr := "02:00:00:00:00"
	return fmt.Sprintf("%s:%x", macAddr, i)
}
