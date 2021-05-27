// +build !js

package dtls

import (
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/pion/dtls/v2/internal/net/dpipe"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/transport/test"
)

func TestContextConfig(t *testing.T) {
	// Limit runtime in case of deadlocks
	lim := test.TimeOut(time.Second * 20)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	addrListen, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Dummy listener
	listen, err := net.ListenUDP("udp", addrListen)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer func() {
		_ = listen.Close()
	}()
	addr := listen.LocalAddr().(*net.UDPAddr)

	cert, err := selfsign.GenerateSelfSigned()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	config := &Config{
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(context.Background(), 40*time.Millisecond)
		},
		Certificates: []tls.Certificate{cert},
	}

	dials := map[string]struct {
		f     func() (func() (net.Conn, error), func())
		order []byte
	}{
		"Dial": {
			f: func() (func() (net.Conn, error), func()) {
				return func() (net.Conn, error) {
						return Dial("udp", addr, config)
					}, func() {
					}
			},
			order: []byte{0, 1, 2},
		},
		"DialWithContext": {
			f: func() (func() (net.Conn, error), func()) {
				ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
				return func() (net.Conn, error) {
						return DialWithContext(ctx, "udp", addr, config)
					}, func() {
						cancel()
					}
			},
			order: []byte{0, 2, 1},
		},
		"Client": {
			f: func() (func() (net.Conn, error), func()) {
				ca, _ := dpipe.Pipe()
				return func() (net.Conn, error) {
						return Client(ca, config)
					}, func() {
						_ = ca.Close()
					}
			},
			order: []byte{0, 1, 2},
		},
		"ClientWithContext": {
			f: func() (func() (net.Conn, error), func()) {
				ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
				ca, _ := dpipe.Pipe()
				return func() (net.Conn, error) {
						return ClientWithContext(ctx, ca, config)
					}, func() {
						cancel()
						_ = ca.Close()
					}
			},
			order: []byte{0, 2, 1},
		},
		"Server": {
			f: func() (func() (net.Conn, error), func()) {
				ca, _ := dpipe.Pipe()
				return func() (net.Conn, error) {
						return Server(ca, config)
					}, func() {
						_ = ca.Close()
					}
			},
			order: []byte{0, 1, 2},
		},
		"ServerWithContext": {
			f: func() (func() (net.Conn, error), func()) {
				ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
				ca, _ := dpipe.Pipe()
				return func() (net.Conn, error) {
						return ServerWithContext(ctx, ca, config)
					}, func() {
						cancel()
						_ = ca.Close()
					}
			},
			order: []byte{0, 2, 1},
		},
	}

	for name, dial := range dials {
		dial := dial
		t.Run(name, func(t *testing.T) {
			done := make(chan struct{})

			go func() {
				d, cancel := dial.f()
				conn, err := d()
				defer cancel()
				if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
					t.Errorf("Client error exp(Temporary network error) failed(%v)", err)
					close(done)
					return
				}
				done <- struct{}{}
				if err == nil {
					_ = conn.Close()
				}
			}()

			var order []byte
			early := time.After(20 * time.Millisecond)
			late := time.After(60 * time.Millisecond)
			func() {
				for len(order) < 3 {
					select {
					case <-early:
						order = append(order, 0)
					case _, ok := <-done:
						if !ok {
							return
						}
						order = append(order, 1)
					case <-late:
						order = append(order, 2)
					}
				}
			}()
			if !bytes.Equal(dial.order, order) {
				t.Errorf("Invalid cancel timing, expected: %v, got: %v", dial.order, order)
			}
		})
	}
}
