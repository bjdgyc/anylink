// +build !js

package dpipe

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"golang.org/x/net/nettest"
)

func TestNetTest(t *testing.T) {
	nettest.TestConn(t, func() (net.Conn, net.Conn, func(), error) {
		ca, cb := Pipe()
		return &closePropagator{ca.(*conn), cb.(*conn)},
			&closePropagator{cb.(*conn), ca.(*conn)},
			func() {
				_ = ca.Close()
				_ = cb.Close()
			}, nil
	})
}

type closePropagator struct {
	*conn
	otherEnd *conn
}

func (c *closePropagator) Close() error {
	close(c.otherEnd.closing)
	return c.conn.Close()
}

func TestPipe(t *testing.T) {
	ca, cb := Pipe()

	testData := []byte{0x01, 0x02}

	for name, cond := range map[string]struct {
		ca net.Conn
		cb net.Conn
	}{
		"AtoB": {ca, cb},
		"BtoA": {cb, ca},
	} {
		c0 := cond.ca
		c1 := cond.cb
		t.Run(name, func(t *testing.T) {
			switch n, err := c0.Write(testData); {
			case err != nil:
				t.Errorf("Unexpected error on Write: %v", err)
			case n != len(testData):
				t.Errorf("Expected to write %d bytes, wrote %d bytes", len(testData), n)
			}

			readData := make([]byte, 4)
			switch n, err := c1.Read(readData); {
			case err != nil:
				t.Errorf("Unexpected error on Write: %v", err)
			case n != len(testData):
				t.Errorf("Expected to read %d bytes, got %d bytes", len(testData), n)
			case !bytes.Equal(testData, readData[0:n]):
				t.Errorf("Expected to read %v, got %v", testData, readData[0:n])
			}
		})
	}

	if err := ca.Close(); err != nil {
		t.Errorf("Unexpected error on Close: %v", err)
	}
	if _, err := ca.Write(testData); !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("Write to closed conn should fail with %v, got %v", io.ErrClosedPipe, err)
	}

	// Other side should be writable.
	if _, err := cb.Write(testData); err != nil {
		t.Errorf("Unexpected error on Write: %v", err)
	}

	readData := make([]byte, 4)
	if _, err := ca.Read(readData); !errors.Is(err, io.EOF) {
		t.Errorf("Read from closed conn should fail with %v, got %v", io.EOF, err)
	}

	// Other side should be readable.
	readDone := make(chan struct{})
	go func() {
		readData := make([]byte, 4)
		if n, err := cb.Read(readData); err == nil {
			t.Errorf("Unexpected data %v was arrived to orphaned conn", readData[:n])
		}
		close(readDone)
	}()
	select {
	case <-readDone:
		t.Errorf("Read should be blocked if the other side is closed")
	case <-time.After(10 * time.Millisecond):
	}
	if err := cb.Close(); err != nil {
		t.Errorf("Unexpected error on Close: %v", err)
	}
}
