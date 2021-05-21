// +build !js

package e2e

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/transport/test"
)

const (
	testMessage   = "Hello World"
	testTimeLimit = 5 * time.Second
	messageRetry  = 200 * time.Millisecond
)

var errServerTimeout = errors.New("waiting on serverReady err: timeout")

func randomPort(t testing.TB) int {
	t.Helper()
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to pickPort: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()
	switch addr := conn.LocalAddr().(type) {
	case *net.UDPAddr:
		return addr.Port
	default:
		t.Fatalf("unknown addr type %T", addr)
		return 0
	}
}

func simpleReadWrite(errChan chan error, outChan chan string, conn io.ReadWriter, messageRecvCount *uint64) {
	go func() {
		buffer := make([]byte, 8192)
		n, err := conn.Read(buffer)
		if err != nil {
			errChan <- err
			return
		}

		outChan <- string(buffer[:n])
		atomic.AddUint64(messageRecvCount, 1)
	}()

	for {
		if atomic.LoadUint64(messageRecvCount) == 2 {
			break
		} else if _, err := conn.Write([]byte(testMessage)); err != nil {
			errChan <- err
			break
		}

		time.Sleep(messageRetry)
	}
}

type comm struct {
	ctx                        context.Context
	clientConfig, serverConfig *dtls.Config
	serverPort                 int
	messageRecvCount           *uint64 // Counter to make sure both sides got a message
	clientMutex                *sync.Mutex
	clientConn                 net.Conn
	serverMutex                *sync.Mutex
	serverConn                 net.Conn
	serverListener             net.Listener
	serverReady                chan struct{}
	errChan                    chan error
	clientChan                 chan string
	serverChan                 chan string
	client                     func(*comm)
	server                     func(*comm)
}

func newComm(ctx context.Context, clientConfig, serverConfig *dtls.Config, serverPort int, server, client func(*comm)) *comm {
	messageRecvCount := uint64(0)
	c := &comm{
		ctx:              ctx,
		clientConfig:     clientConfig,
		serverConfig:     serverConfig,
		serverPort:       serverPort,
		messageRecvCount: &messageRecvCount,
		clientMutex:      &sync.Mutex{},
		serverMutex:      &sync.Mutex{},
		serverReady:      make(chan struct{}),
		errChan:          make(chan error),
		clientChan:       make(chan string),
		serverChan:       make(chan string),
		server:           server,
		client:           client,
	}
	return c
}

func (c *comm) assert(t *testing.T) {
	// DTLS Client
	go c.client(c)

	// DTLS Server
	go c.server(c)

	defer func() {
		if c.clientConn != nil {
			if err := c.clientConn.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if c.serverConn != nil {
			if err := c.serverConn.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if c.serverListener != nil {
			if err := c.serverListener.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}()

	func() {
		seenClient, seenServer := false, false
		for {
			select {
			case err := <-c.errChan:
				t.Fatal(err)
			case <-time.After(testTimeLimit):
				t.Fatalf("Test timeout, seenClient %t seenServer %t", seenClient, seenServer)
			case clientMsg := <-c.clientChan:
				if clientMsg != testMessage {
					t.Fatalf("clientMsg does not equal test message: %s %s", clientMsg, testMessage)
				}

				seenClient = true
				if seenClient && seenServer {
					return
				}
			case serverMsg := <-c.serverChan:
				if serverMsg != testMessage {
					t.Fatalf("serverMsg does not equal test message: %s %s", serverMsg, testMessage)
				}

				seenServer = true
				if seenClient && seenServer {
					return
				}
			}
		}
	}()
}

func clientPion(c *comm) {
	select {
	case <-c.serverReady:
		// OK
	case <-time.After(time.Second):
		c.errChan <- errServerTimeout
	}

	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

	var err error
	c.clientConn, err = dtls.DialWithContext(c.ctx, "udp",
		&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: c.serverPort},
		c.clientConfig,
	)
	if err != nil {
		c.errChan <- err
		return
	}

	simpleReadWrite(c.errChan, c.clientChan, c.clientConn, c.messageRecvCount)
}

func serverPion(c *comm) {
	c.serverMutex.Lock()
	defer c.serverMutex.Unlock()

	var err error
	c.serverListener, err = dtls.Listen("udp",
		&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: c.serverPort},
		c.serverConfig,
	)
	if err != nil {
		c.errChan <- err
		return
	}
	c.serverReady <- struct{}{}
	c.serverConn, err = c.serverListener.Accept()
	if err != nil {
		c.errChan <- err
		return
	}

	simpleReadWrite(c.errChan, c.serverChan, c.serverConn, c.messageRecvCount)
}

/*
  Simple DTLS Client/Server can communicate
    - Assert that you can send messages both ways
	- Assert that Close() on both ends work
	- Assert that no Goroutines are leaked
*/
func testPionE2ESimple(t *testing.T, server, client func(*comm)) {
	lim := test.TimeOut(time.Second * 30)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	for _, cipherSuite := range []dtls.CipherSuiteID{
		dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		dtls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	} {
		cipherSuite := cipherSuite
		t.Run(cipherSuite.String(), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			cert, err := selfsign.GenerateSelfSignedWithDNS("localhost")
			if err != nil {
				t.Fatal(err)
			}

			cfg := &dtls.Config{
				Certificates:       []tls.Certificate{cert},
				CipherSuites:       []dtls.CipherSuiteID{cipherSuite},
				InsecureSkipVerify: true,
			}
			serverPort := randomPort(t)
			comm := newComm(ctx, cfg, cfg, serverPort, server, client)
			comm.assert(t)
		})
	}
}

func testPionE2ESimplePSK(t *testing.T, server, client func(*comm)) {
	lim := test.TimeOut(time.Second * 30)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	for _, cipherSuite := range []dtls.CipherSuiteID{
		dtls.TLS_PSK_WITH_AES_128_CCM,
		dtls.TLS_PSK_WITH_AES_128_CCM_8,
		dtls.TLS_PSK_WITH_AES_128_GCM_SHA256,
	} {
		cipherSuite := cipherSuite
		t.Run(cipherSuite.String(), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			cfg := &dtls.Config{
				PSK: func(hint []byte) ([]byte, error) {
					return []byte{0xAB, 0xC1, 0x23}, nil
				},
				PSKIdentityHint: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				CipherSuites:    []dtls.CipherSuiteID{cipherSuite},
			}
			serverPort := randomPort(t)
			comm := newComm(ctx, cfg, cfg, serverPort, server, client)
			comm.assert(t)
		})
	}
}

func testPionE2EMTUs(t *testing.T, server, client func(*comm)) {
	lim := test.TimeOut(time.Second * 30)
	defer lim.Stop()

	report := test.CheckRoutines(t)
	defer report()

	for _, mtu := range []int{
		10000,
		1000,
		100,
	} {
		mtu := mtu
		t.Run(fmt.Sprintf("MTU%d", mtu), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cert, err := selfsign.GenerateSelfSignedWithDNS("localhost")
			if err != nil {
				t.Fatal(err)
			}

			cfg := &dtls.Config{
				Certificates:       []tls.Certificate{cert},
				CipherSuites:       []dtls.CipherSuiteID{dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
				InsecureSkipVerify: true,
				MTU:                mtu,
			}
			serverPort := randomPort(t)
			comm := newComm(ctx, cfg, cfg, serverPort, server, client)
			comm.assert(t)
		})
	}
}

func TestPionE2ESimple(t *testing.T) {
	testPionE2ESimple(t, serverPion, clientPion)
}

func TestPionE2ESimplePSK(t *testing.T) {
	testPionE2ESimplePSK(t, serverPion, clientPion)
}

func TestPionE2EMTUs(t *testing.T) {
	testPionE2EMTUs(t, serverPion, clientPion)
}
