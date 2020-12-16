// copy from: https://github.com/armon/go-proxyproto/blob/master/protocol_test.go
package proxyproto

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

const (
	goodAddr = "127.0.0.1"
	badAddr  = "127.0.0.2"
	errAddr  = "9999.0.0.2"
)

var (
	checkAddr string
)

func TestPassthrough(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pl := &Listener{Listener: l}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		conn.Write([]byte("ping"))
		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !bytes.Equal(recv, []byte("pong")) {
			t.Fatalf("bad: %v", recv)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestTimeout(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	clientWriteDelay := 200 * time.Millisecond
	proxyHeaderTimeout := 50 * time.Millisecond
	pl := &Listener{Listener: l, ProxyHeaderTimeout: proxyHeaderTimeout}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		// Do not send data for a while
		time.Sleep(clientWriteDelay)

		conn.Write([]byte("ping"))
		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !bytes.Equal(recv, []byte("pong")) {
			t.Fatalf("bad: %v", recv)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	// Check the remote addr is the original 127.0.0.1
	remoteAddrStartTime := time.Now()
	addr := conn.RemoteAddr().(*net.TCPAddr)
	if addr.IP.String() != "127.0.0.1" {
		t.Fatalf("bad: %v", addr)
	}
	remoteAddrDuration := time.Since(remoteAddrStartTime)

	// Check RemoteAddr() call did timeout
	if remoteAddrDuration >= clientWriteDelay {
		t.Fatalf("RemoteAddr() took longer than the specified timeout: %v < %v", proxyHeaderTimeout, remoteAddrDuration)
	}

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestParse_ipv4(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pl := &Listener{Listener: l}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		// Write out the header!
		header := "PROXY TCP4 10.1.1.1 20.2.2.2 1000 2000\r\n"
		conn.Write([]byte(header))

		conn.Write([]byte("ping"))
		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !bytes.Equal(recv, []byte("pong")) {
			t.Fatalf("bad: %v", recv)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the remote addr
	addr := conn.RemoteAddr().(*net.TCPAddr)
	if addr.IP.String() != "10.1.1.1" {
		t.Fatalf("bad: %v", addr)
	}
	if addr.Port != 1000 {
		t.Fatalf("bad: %v", addr)
	}
}

func TestParse_ipv6(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pl := &Listener{Listener: l}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		// Write out the header!
		header := "PROXY TCP6 ffff::ffff ffff::ffff 1000 2000\r\n"
		conn.Write([]byte(header))

		conn.Write([]byte("ping"))
		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !bytes.Equal(recv, []byte("pong")) {
			t.Fatalf("bad: %v", recv)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the remote addr
	addr := conn.RemoteAddr().(*net.TCPAddr)
	if addr.IP.String() != "ffff::ffff" {
		t.Fatalf("bad: %v", addr)
	}
	if addr.Port != 1000 {
		t.Fatalf("bad: %v", addr)
	}
}

func TestParse_Unknown(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pl := &Listener{Listener: l, UnknownOK: true}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		// Write out the header!
		header := "PROXY UNKNOWN\r\n"
		conn.Write([]byte(header))

		conn.Write([]byte("ping"))
		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !bytes.Equal(recv, []byte("pong")) {
			t.Fatalf("bad: %v", recv)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}

}

func TestParse_BadHeader(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	pl := &Listener{Listener: l}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		// Write out the header!
		header := "PROXY TCP4 what 127.0.0.1 1000 2000\r\n"
		conn.Write([]byte(header))

		conn.Write([]byte("ping"))

		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err == nil {
			t.Fatalf("err: %v", err)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	// Check the remote addr, should be the local addr
	addr := conn.RemoteAddr().(*net.TCPAddr)
	if addr.IP.String() != "127.0.0.1" {
		t.Fatalf("bad: %v", addr)
	}

	// Read should fail
	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err == nil {
		t.Fatalf("err: %v", err)
	}
}

func TestParse_ipv4_checkfunc(t *testing.T) {
	checkAddr = goodAddr
	testParse_ipv4_checkfunc(t)
	checkAddr = badAddr
	testParse_ipv4_checkfunc(t)
	checkAddr = errAddr
	testParse_ipv4_checkfunc(t)
}

func testParse_ipv4_checkfunc(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	checkFunc := func(addr net.Addr) (bool, error) {
		tcpAddr := addr.(*net.TCPAddr)
		if tcpAddr.IP.String() == checkAddr {
			return true, nil
		}
		return false, nil
	}

	pl := &Listener{Listener: l, SourceCheck: checkFunc}

	go func() {
		conn, err := net.Dial("tcp", pl.Addr().String())
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		// Write out the header!
		header := "PROXY TCP4 10.1.1.1 20.2.2.2 1000 2000\r\n"
		conn.Write([]byte(header))

		conn.Write([]byte("ping"))
		recv := make([]byte, 4)
		_, err = conn.Read(recv)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !bytes.Equal(recv, []byte("pong")) {
			t.Fatalf("bad: %v", recv)
		}
	}()

	conn, err := pl.Accept()
	if err != nil {
		if checkAddr == badAddr {
			return
		}
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the remote addr
	addr := conn.RemoteAddr().(*net.TCPAddr)
	switch checkAddr {
	case goodAddr:
		if addr.IP.String() != "10.1.1.1" {
			t.Fatalf("bad: %v", addr)
		}
		if addr.Port != 1000 {
			t.Fatalf("bad: %v", addr)
		}
	case badAddr:
		if addr.IP.String() != "127.0.0.1" {
			t.Fatalf("bad: %v", addr)
		}
		if addr.Port == 1000 {
			t.Fatalf("bad: %v", addr)
		}
	}
}

type testConn struct {
	readFromCalledWith io.Reader
	net.Conn           // nil; crash on any unexpected use
}

func (c *testConn) ReadFrom(r io.Reader) (int64, error) {
	c.readFromCalledWith = r
	return 0, nil
}
func (c *testConn) Write(p []byte) (int, error) {
	return len(p), nil
}
func (c *testConn) Read(p []byte) (int, error) {
	return 1, nil
}

func TestCopyToWrappedConnection(t *testing.T) {
	innerConn := &testConn{}
	wrappedConn := NewConn(innerConn, 0)
	dummySrc := &testConn{}

	io.Copy(wrappedConn, dummySrc)
	if innerConn.readFromCalledWith != dummySrc {
		t.Error("Expected io.Copy to delegate to ReadFrom function of inner destination connection")
	}
}

func TestCopyFromWrappedConnection(t *testing.T) {
	wrappedConn := NewConn(&testConn{}, 0)
	dummyDst := &testConn{}

	io.Copy(dummyDst, wrappedConn)
	if dummyDst.readFromCalledWith != wrappedConn.conn {
		t.Errorf("Expected io.Copy to pass inner source connection to ReadFrom method of destination")
	}
}

func TestCopyFromWrappedConnectionToWrappedConnection(t *testing.T) {
	innerConn1 := &testConn{}
	wrappedConn1 := NewConn(innerConn1, 0)
	innerConn2 := &testConn{}
	wrappedConn2 := NewConn(innerConn2, 0)

	io.Copy(wrappedConn1, wrappedConn2)
	if innerConn1.readFromCalledWith != innerConn2 {
		t.Errorf("Expected io.Copy to pass inner source connection to ReadFrom of inner destination connection")
	}

}
