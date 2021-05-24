// Package dpipe provides the pipe works like datagram protocol on memory.
package dpipe

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pion/transport/deadline"
)

// Pipe creates pair of non-stream conn on memory.
// Close of the one end doesn't make effect to the other end.
func Pipe() (net.Conn, net.Conn) {
	ch0 := make(chan []byte, 1000)
	ch1 := make(chan []byte, 1000)
	return &conn{
			rCh:           ch0,
			wCh:           ch1,
			closed:        make(chan struct{}),
			closing:       make(chan struct{}),
			readDeadline:  deadline.New(),
			writeDeadline: deadline.New(),
		}, &conn{
			rCh:           ch1,
			wCh:           ch0,
			closed:        make(chan struct{}),
			closing:       make(chan struct{}),
			readDeadline:  deadline.New(),
			writeDeadline: deadline.New(),
		}
}

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return ":1" }

type conn struct {
	rCh       chan []byte
	wCh       chan []byte
	closed    chan struct{}
	closing   chan struct{}
	closeOnce sync.Once

	readDeadline  *deadline.Deadline
	writeDeadline *deadline.Deadline
}

func (*conn) LocalAddr() net.Addr  { return pipeAddr{} }
func (*conn) RemoteAddr() net.Addr { return pipeAddr{} }

func (c *conn) SetDeadline(t time.Time) error {
	c.readDeadline.Set(t)
	c.writeDeadline.Set(t)
	return nil
}

func (c *conn) SetReadDeadline(t time.Time) error {
	c.readDeadline.Set(t)
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline.Set(t)
	return nil
}

func (c *conn) Read(data []byte) (n int, err error) {
	select {
	case <-c.closed:
		return 0, io.EOF
	case <-c.closing:
		if len(c.rCh) == 0 {
			return 0, io.EOF
		}
	case <-c.readDeadline.Done():
		return 0, context.DeadlineExceeded
	default:
	}

	for {
		select {
		case d := <-c.rCh:
			if len(d) <= len(data) {
				copy(data, d)
				return len(d), nil
			}
			copy(data, d[:len(data)])
			return len(data), nil
		case <-c.closed:
			return 0, io.EOF
		case <-c.closing:
			if len(c.rCh) == 0 {
				return 0, io.EOF
			}
		case <-c.readDeadline.Done():
			return 0, context.DeadlineExceeded
		}
	}
}

func (c *conn) cleanWriteBuffer() {
	for {
		select {
		case <-c.wCh:
		default:
			return
		}
	}
}

func (c *conn) Write(data []byte) (n int, err error) {
	select {
	case <-c.closed:
		return 0, io.ErrClosedPipe
	case <-c.writeDeadline.Done():
		c.cleanWriteBuffer()
		return 0, context.DeadlineExceeded
	default:
	}

	cData := make([]byte, len(data))
	copy(cData, data)

	select {
	case <-c.closed:
		return 0, io.ErrClosedPipe
	case <-c.writeDeadline.Done():
		c.cleanWriteBuffer()
		return 0, context.DeadlineExceeded
	case c.wCh <- cData:
		return len(cData), nil
	}
}

func (c *conn) Close() error {
	c.closeOnce.Do(func() {
		close(c.closed)
	})
	return nil
}
