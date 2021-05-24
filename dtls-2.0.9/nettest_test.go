// +build !js

package dtls

import (
	"net"
	"testing"
	"time"

	"github.com/pion/transport/test"
	"golang.org/x/net/nettest"
)

func TestNetTest(t *testing.T) {
	lim := test.TimeOut(time.Minute*1 + time.Second*10)
	defer lim.Stop()

	nettest.TestConn(t, func() (c1, c2 net.Conn, stop func(), err error) {
		c1, c2, err = pipeMemory()
		if err != nil {
			return nil, nil, nil, err
		}
		stop = func() {
			_ = c1.Close()
			_ = c2.Close()
		}
		return
	})
}
