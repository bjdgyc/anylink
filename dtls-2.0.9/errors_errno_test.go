// +build aix darwin dragonfly freebsd linux nacl nacljs netbsd openbsd solaris windows

// For systems having syscall.Errno.
// The build target must be same as errors_errno.go.

package dtls

import (
	"net"
	"testing"
)

func TestErrorsTemporary(t *testing.T) {
	addrListen, errListen := net.ResolveUDPAddr("udp", "localhost:0")
	if errListen != nil {
		t.Fatalf("Unexpected error: %v", errListen)
	}
	// Server is not listening.
	conn, errDial := net.DialUDP("udp", nil, addrListen)
	if errDial != nil {
		t.Fatalf("Unexpected error: %v", errDial)
	}

	_, _ = conn.Write([]byte{0x00}) // trigger
	_, err := conn.Read(make([]byte, 10))
	_ = conn.Close()

	if err == nil {
		t.Skip("ECONNREFUSED is not set by system")
	}
	ne, ok := netError(err).(net.Error)
	if !ok {
		t.Fatalf("netError must return net.Error")
	}
	if ne.Timeout() {
		t.Errorf("%v must not be timeout error", err)
	}
	if !ne.Temporary() {
		t.Errorf("%v must be temporary error", err)
	}
}
