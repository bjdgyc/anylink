package dtls

import (
	"context"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pion/dtls/v2/internal/net/dpipe"
	"github.com/pion/transport/test"
)

func TestReplayProtection(t *testing.T) {
	// Limit runtime in case of deadlocks
	lim := test.TimeOut(5 * time.Second)
	defer lim.Stop()

	// Check for leaking routines
	report := test.CheckRoutines(t)
	defer report()

	c0, c1 := dpipe.Pipe()
	c2, c3 := dpipe.Pipe()
	conn := []net.Conn{c0, c1, c2, c3}

	var wgRoutines sync.WaitGroup
	var cntReplays int32 = 1

	ctxReplayDone, replayDone := context.WithCancel(context.Background())

	replaySendDone := func() {
		cnt := atomic.AddInt32(&cntReplays, -1)
		if cnt == 0 {
			replayDone()
		}
	}

	replayer := func(ca, cb net.Conn) {
		defer wgRoutines.Done()
		// Man in the middle
		for {
			b := make([]byte, 2048)
			n, rerr := ca.Read(b)
			if rerr != nil {
				return
			}
			if _, werr := cb.Write(b[:n]); werr != nil {
				t.Error(werr)
				return
			}

			atomic.AddInt32(&cntReplays, 1)
			go func() {
				defer replaySendDone()
				// Replay bit later
				time.Sleep(time.Millisecond)
				if _, werr := cb.Write(b[:n]); werr != nil {
					t.Error(werr)
				}
			}()
		}
	}
	wgRoutines.Add(2)
	go replayer(conn[1], conn[2])
	go replayer(conn[2], conn[1])

	ca, cb, err := pipeConn(conn[0], conn[3])
	if err != nil {
		t.Fatal(err)
	}

	const numMsgs = 10

	var received [2][][]byte
	for i, c := range []net.Conn{ca, cb} {
		i := i
		c := c
		wgRoutines.Add(1)
		atomic.AddInt32(&cntReplays, 1) // Keep locked until the final message
		var lastMsgDone sync.Once
		go func() {
			defer wgRoutines.Done()
			for {
				b := make([]byte, 2048)
				n, rerr := c.Read(b)
				if rerr != nil {
					return
				}
				received[i] = append(received[i], b[:n])
				if b[0] == numMsgs-1 {
					// Final message received
					lastMsgDone.Do(func() {
						defer replaySendDone()
					})
				}
			}
		}()
	}

	var sent [][]byte
	for i := 0; i < numMsgs; i++ {
		data := []byte{byte(i)}
		sent = append(sent, data)
		if _, werr := ca.Write(data); werr != nil {
			t.Error(werr)
			return
		}
		if _, werr := cb.Write(data); werr != nil {
			t.Error(werr)
			return
		}
	}

	replaySendDone()
	<-ctxReplayDone.Done()
	time.Sleep(10 * time.Millisecond) // Ensure all replayed packets are sent

	for i := 0; i < 4; i++ {
		if err := conn[i].Close(); err != nil {
			t.Error(err)
		}
	}
	if err := ca.Close(); err != nil {
		t.Error(err)
	}
	if err := cb.Close(); err != nil {
		t.Error(err)
	}
	wgRoutines.Wait()

	for _, r := range received {
		if !reflect.DeepEqual(sent, r) {
			t.Errorf("Received data differs, expected: %v, got: %v", sent, r)
		}
	}
}
