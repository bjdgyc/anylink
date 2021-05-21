package dtls

import (
	"bytes"
	"context"
	"crypto/tls"
	"sync"
	"testing"
	"time"

	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/dtls/v2/pkg/crypto/signaturehash"
	"github.com/pion/dtls/v2/pkg/protocol/alert"
	"github.com/pion/dtls/v2/pkg/protocol/handshake"
	"github.com/pion/dtls/v2/pkg/protocol/recordlayer"
	"github.com/pion/logging"
	"github.com/pion/transport/test"
)

const nonZeroRetransmitInterval = 100 * time.Millisecond

// Test that writes to the key log are in the correct format and only applies
// when a key log writer is given.
func TestWriteKeyLog(t *testing.T) {
	var buf bytes.Buffer
	cfg := handshakeConfig{
		keyLogWriter: &buf,
	}
	cfg.writeKeyLog("LABEL", []byte{0xAA, 0xBB, 0xCC}, []byte{0xDD, 0xEE, 0xFF})

	// Secrets follow the format <Label> <space> <ClientRandom> <space> <Secret>
	// https://developer.mozilla.org/en-US/docs/Mozilla/Projects/NSS/Key_Log_Format
	want := "LABEL aabbcc ddeeff\n"
	if buf.String() != want {
		t.Fatalf("Got %s want %s", buf.String(), want)
	}

	// no key log writer = no writes
	cfg = handshakeConfig{}
	cfg.writeKeyLog("LABEL", []byte{0xAA, 0xBB, 0xCC}, []byte{0xDD, 0xEE, 0xFF})
}

func TestHandshaker(t *testing.T) {
	// Check for leaking routines
	report := test.CheckRoutines(t)
	defer report()

	loggerFactory := logging.NewDefaultLoggerFactory()
	logger := loggerFactory.NewLogger("dtls")

	cipherSuites, err := parseCipherSuites(nil, nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	clientCert, err := selfsign.GenerateSelfSigned()
	if err != nil {
		t.Fatal(err)
	}

	genFilters := map[string]func() (packetFilter, packetFilter, func(t *testing.T)){
		"PassThrough": func() (packetFilter, packetFilter, func(t *testing.T)) {
			return nil, nil, nil
		},
		"HelloVerifyRequestLost": func() (packetFilter, packetFilter, func(t *testing.T)) {
			var (
				cntHelloVerifyRequest  = 0
				cntClientHelloNoCookie = 0
			)
			const helloVerifyDrop = 5
			return func(p *packet) bool {
					h, ok := p.record.Content.(*handshake.Handshake)
					if !ok {
						return true
					}
					if hmch, ok := h.Message.(*handshake.MessageClientHello); ok {
						if len(hmch.Cookie) == 0 {
							cntClientHelloNoCookie++
						}
					}
					return true
				},
				func(p *packet) bool {
					h, ok := p.record.Content.(*handshake.Handshake)
					if !ok {
						return true
					}
					if _, ok := h.Message.(*handshake.MessageHelloVerifyRequest); ok {
						cntHelloVerifyRequest++
						return cntHelloVerifyRequest > helloVerifyDrop
					}
					return true
				},
				func(t *testing.T) {
					if cntHelloVerifyRequest != helloVerifyDrop+1 {
						t.Errorf("Number of HelloVerifyRequest retransmit is wrong, expected: %d times, got: %d times", helloVerifyDrop+1, cntHelloVerifyRequest)
					}
					if cntClientHelloNoCookie != cntHelloVerifyRequest {
						t.Errorf(
							"HelloVerifyRequest must be triggered only by ClientHello, but HelloVerifyRequest was sent %d times and ClientHello was sent %d times",
							cntHelloVerifyRequest, cntClientHelloNoCookie,
						)
					}
				}
		},
	}

	for name, filters := range genFilters {
		f1, f2, report := filters()
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if report != nil {
				defer report(t)
			}

			ca, cb := flightTestPipe(ctx, f1, f2)
			ca.state.isClient = true

			var wg sync.WaitGroup
			wg.Add(2)

			ctxCliFinished, cancelCli := context.WithCancel(ctx)
			ctxSrvFinished, cancelSrv := context.WithCancel(ctx)
			go func() {
				defer wg.Done()
				cfg := &handshakeConfig{
					localCipherSuites:     cipherSuites,
					localCertificates:     []tls.Certificate{clientCert},
					localSignatureSchemes: signaturehash.Algorithms(),
					insecureSkipVerify:    true,
					log:                   logger,
					onFlightState: func(f flightVal, s handshakeState) {
						if s == handshakeFinished {
							cancelCli()
						}
					},
					retransmitInterval: nonZeroRetransmitInterval,
				}

				fsm := newHandshakeFSM(&ca.state, ca.handshakeCache, cfg, flight1)
				switch err := fsm.Run(ctx, ca, handshakePreparing); err {
				case context.Canceled:
				case context.DeadlineExceeded:
					t.Error("Timeout")
				default:
					t.Error(err)
				}
			}()

			go func() {
				defer wg.Done()
				cfg := &handshakeConfig{
					localCipherSuites:     cipherSuites,
					localCertificates:     []tls.Certificate{clientCert},
					localSignatureSchemes: signaturehash.Algorithms(),
					insecureSkipVerify:    true,
					log:                   logger,
					onFlightState: func(f flightVal, s handshakeState) {
						if s == handshakeFinished {
							cancelSrv()
						}
					},
					retransmitInterval: nonZeroRetransmitInterval,
				}

				fsm := newHandshakeFSM(&cb.state, cb.handshakeCache, cfg, flight0)
				switch err := fsm.Run(ctx, cb, handshakePreparing); err {
				case context.Canceled:
				case context.DeadlineExceeded:
					t.Error("Timeout")
				default:
					t.Error(err)
				}
			}()

			<-ctxCliFinished.Done()
			<-ctxSrvFinished.Done()

			cancel()
			wg.Wait()
		})
	}
}

type packetFilter func(*packet) bool

func flightTestPipe(ctx context.Context, filter1 packetFilter, filter2 packetFilter) (*flightTestConn, *flightTestConn) {
	ca := newHandshakeCache()
	cb := newHandshakeCache()
	chA := make(chan chan struct{})
	chB := make(chan chan struct{})
	return &flightTestConn{
			handshakeCache: ca,
			otherEndCache:  cb,
			recv:           chA,
			otherEndRecv:   chB,
			done:           ctx.Done(),
			filter:         filter1,
		}, &flightTestConn{
			handshakeCache: cb,
			otherEndCache:  ca,
			recv:           chB,
			otherEndRecv:   chA,
			done:           ctx.Done(),
			filter:         filter2,
		}
}

type flightTestConn struct {
	state          State
	handshakeCache *handshakeCache
	recv           chan chan struct{}
	done           <-chan struct{}
	epoch          uint16

	filter packetFilter

	otherEndCache *handshakeCache
	otherEndRecv  chan chan struct{}
}

func (c *flightTestConn) recvHandshake() <-chan chan struct{} {
	return c.recv
}

func (c *flightTestConn) setLocalEpoch(epoch uint16) {
	c.epoch = epoch
}

func (c *flightTestConn) notify(ctx context.Context, level alert.Level, desc alert.Description) error {
	return nil
}

func (c *flightTestConn) writePackets(ctx context.Context, pkts []*packet) error {
	for _, p := range pkts {
		if c.filter != nil && !c.filter(p) {
			continue
		}
		if h, ok := p.record.Content.(*handshake.Handshake); ok {
			handshakeRaw, err := p.record.Marshal()
			if err != nil {
				return err
			}

			c.handshakeCache.push(handshakeRaw[recordlayer.HeaderSize:], p.record.Header.Epoch, h.Header.MessageSequence, h.Header.Type, c.state.isClient)

			content, err := h.Message.Marshal()
			if err != nil {
				return err
			}
			h.Header.Length = uint32(len(content))
			h.Header.FragmentLength = uint32(len(content))
			hdr, err := h.Header.Marshal()
			if err != nil {
				return err
			}
			c.otherEndCache.push(
				append(hdr, content...), p.record.Header.Epoch, h.Header.MessageSequence, h.Header.Type, c.state.isClient)
		}
	}
	go func() {
		select {
		case c.otherEndRecv <- make(chan struct{}):
		case <-c.done:
		}
	}()

	// Avoid deadlock on JS/WASM environment due to context switch problem.
	time.Sleep(10 * time.Millisecond)

	return nil
}

func (c *flightTestConn) handleQueuedPackets(ctx context.Context) error {
	return nil
}
