package common

import (
	"fmt"
	"net"
	"testing"
)

func TestAcquireIp(t *testing.T) {
	ServerCfg.Ipv4Network = "192.168.1.0"
	ServerCfg.Ipv4Netmask = "255.255.255.0"
	macIps = map[string]*MacIp{}
	initIpPool()

	var ip net.IP

	for i := 2; i <= 100; i++ {
		ip = AcquireIp(fmt.Sprintf("mac-%d", i))
	}
	ip = AcquireIp(fmt.Sprintf("mac-new"))
	AssertTrue(t, ip.Equal(net.IPv4(192, 168, 1, 101)))
	for i := 102; i <= 254; i++ {
		ip = AcquireIp(fmt.Sprintf("mac-%d", i))
	}
	ip = AcquireIp(fmt.Sprintf("mac-nil"))
	AssertTrue(t, ip == nil)
}

func TestReleaseIp(t *testing.T) {
	ServerCfg.Ipv4Network = "192.168.1.0"
	ServerCfg.Ipv4Netmask = "255.255.255.0"
	macIps = map[string]*MacIp{}
	initIpPool()

	var ip net.IP

	// 分配完所有数据
	for i := 2; i <= 254; i++ {
		ip = AcquireIp(fmt.Sprintf("mac-%d", i))
	}

	ip = AcquireIp(fmt.Sprintf("mac-more"))
	AssertTrue(t, ip == nil)

	ReleaseIp(net.IPv4(192, 168, 1, 123), "mac-123")
	ReleaseIp(net.IPv4(192, 168, 1, 100), "mac-100")
	ip = AcquireIp(fmt.Sprintf("mac-new"))
	// 最早过期的ip
	AssertTrue(t, ip.Equal(net.IPv4(192, 168, 1, 123)))
}
