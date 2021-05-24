package dtls

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/pion/dtls/v2/internal/net/dpipe"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/logging"
	"github.com/pion/transport/test"
)

func TestSimpleReadWrite(t *testing.T) {
	report := test.CheckRoutines(t)
	defer report()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ca, cb := dpipe.Pipe()
	certificate, err := selfsign.GenerateSelfSigned()
	if err != nil {
		t.Fatal(err)
	}
	gotHello := make(chan struct{})

	go func() {
		server, sErr := testServer(ctx, cb, &Config{
			Certificates:  []tls.Certificate{certificate},
			LoggerFactory: logging.NewDefaultLoggerFactory(),
		}, false)
		if sErr != nil {
			t.Error(sErr)
			return
		}
		buf := make([]byte, 1024)
		if _, sErr = server.Read(buf); sErr != nil {
			t.Error(sErr)
		}
		gotHello <- struct{}{}
		if sErr = server.Close(); sErr != nil {
			t.Error(sErr)
		}
	}()

	client, err := testClient(ctx, ca, &Config{
		LoggerFactory:      logging.NewDefaultLoggerFactory(),
		InsecureSkipVerify: true,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.Write([]byte("hello")); err != nil {
		t.Error(err)
	}
	select {
	case <-gotHello:
		// OK
	case <-time.After(time.Second * 5):
		t.Error("timeout")
	}

	if err = client.Close(); err != nil {
		t.Error(err)
	}
}

func benchmarkConn(b *testing.B, n int64) {
	b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
		ctx := context.Background()

		ca, cb := dpipe.Pipe()
		certificate, err := selfsign.GenerateSelfSigned()
		server := make(chan *Conn)
		go func() {
			s, sErr := testServer(ctx, cb, &Config{
				Certificates: []tls.Certificate{certificate},
			}, false)
			if err != nil {
				b.Error(sErr)
				return
			}
			server <- s
		}()
		if err != nil {
			b.Fatal(err)
		}
		hw := make([]byte, n)
		b.ReportAllocs()
		b.SetBytes(int64(len(hw)))
		go func() {
			client, cErr := testClient(ctx, ca, &Config{InsecureSkipVerify: true}, false)
			if cErr != nil {
				b.Error(err)
			}
			for {
				if _, cErr = client.Write(hw); cErr != nil {
					b.Error(err)
				}
			}
		}()
		s := <-server
		buf := make([]byte, 2048)
		for i := 0; i < b.N; i++ {
			if _, err = s.Read(buf); err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkConnReadWrite(b *testing.B) {
	for _, n := range []int64{16, 128, 512, 1024, 2048} {
		benchmarkConn(b, n)
	}
}
