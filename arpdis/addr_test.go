package arpdis

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {
	assert := assert.New(t)
	ip := net.IPv4(192, 168, 10, 2)
	hw, _ := net.ParseMAC("08:00:27:a0:17:42")
	now := time.Now()
	addr1 := &Addr{
		IP:           ip,
		HardwareAddr: hw,
		Type:         TypeStatic,
		disTime:      now,
	}
	Add(addr1)
	addr2 := Lookup(ip, true)
	assert.Equal(addr1, addr2)
	addr3 := &Addr{
		IP:           ip,
		HardwareAddr: hw,
		Type:         TypeNormal,
		disTime:      now,
	}
	Add(addr3)
	addr4 := Lookup(ip, true)
	// 静态地址只能设置一次
	assert.NotEqual(addr3, addr4)
}
