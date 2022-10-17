package handler

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/logging"
)

func startDtls() {
	if !base.Cfg.ServerDTLS {
		return
	}

	certificate, err := selfsign.GenerateSelfSigned()
	if err != nil {
		panic(err)
	}
	logf := logging.NewDefaultLoggerFactory()
	logf.Writer = base.GetBaseLw()
	// logf.DefaultLogLevel = logging.LogLevelTrace
	logf.DefaultLogLevel = logging.LogLevelInfo

	// https://github.com/pion/dtls/pull/369
	sessStore := &sessionStore{}

	config := &dtls.Config{
		Certificates:         []tls.Certificate{certificate},
		InsecureSkipVerify:   true,
		ExtendedMasterSecret: dtls.DisableExtendedMasterSecret,
		CipherSuites:         []dtls.CipherSuiteID{dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		LoggerFactory:        logf,
		MTU:                  BufferSize,
		SessionStore:         sessStore,
		ConnectContextMaker: func() (context.Context, func()) {
			return context.WithTimeout(context.Background(), 5*time.Second)
		},
	}

	addr, err := net.ResolveUDPAddr("udp", base.Cfg.ServerDTLSAddr)
	if err != nil {
		panic(err)
	}
	ln, err := dtls.Listen("udp", addr, config)
	if err != nil {
		panic(err)
	}

	base.Info("listen DTLS server", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			base.Error("DTLS Accept error", err)
			continue
		}

		go func() {
			// time.Sleep(1 * time.Second)
			cc := conn.(*dtls.Conn)
			sessid := hex.EncodeToString(cc.ConnectionState().SessionID)
			sess := sessdata.Dtls2Sess(sessid)
			LinkDtls(conn, sess.CSess)
		}()
	}
}

// https://github.com/pion/dtls/blob/master/session.go
type sessionStore struct{}

func (ms *sessionStore) Set(key []byte, s dtls.Session) error {
	return nil
}

func (ms *sessionStore) Get(key []byte) (dtls.Session, error) {
	k := hex.EncodeToString(key)
	secret := sessdata.Dtls2MasterSecret(k)
	if secret == "" {
		return dtls.Session{}, errors.New("Dtls2MasterSecret is nil")
	}

	masterSecret, _ := hex.DecodeString(secret)
	return dtls.Session{ID: key, Secret: masterSecret}, nil
}

func (ms *sessionStore) Del(key []byte) error {
	return nil
}
