package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"net"
	"strings"
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

	// rsa 兼容 open connect
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	certificate, err := selfsign.SelfSign(priv)
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
		ExtendedMasterSecret: dtls.DisableExtendedMasterSecret,
		CipherSuites: func() []dtls.CipherSuiteID {
			var cs = []dtls.CipherSuiteID{}
			for _, vv := range dtlsCipherSuites {
				cs = append(cs, vv)
			}
			return cs
		}(),
		LoggerFactory: logf,
		MTU:           BufferSize,
		SessionStore:  sessStore,
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
			did := hex.EncodeToString(cc.ConnectionState().SessionID)
			cSess := sessdata.Dtls2CSess(did)
			if cSess == nil {
				conn.Close()
				return
			}
			LinkDtls(conn, cSess)
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

// 客户端和服务端映射 X-DTLS12-CipherSuite
var dtlsCipherSuites = map[string]dtls.CipherSuiteID{
	// "ECDHE-ECDSA-AES256-GCM-SHA384": dtls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	// "ECDHE-ECDSA-AES128-GCM-SHA256": dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"ECDHE-RSA-AES256-GCM-SHA384": dtls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-RSA-AES128-GCM-SHA256": dtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

func checkDtls12Ciphersuite(ciphersuite string) string {
	csArr := strings.Split(ciphersuite, ":")

	for _, v := range csArr {
		if _, ok := dtlsCipherSuites[v]; ok {
			return v
		}
	}
	// 返回默认值
	return "ECDHE-RSA-AES128-GCM-SHA256"
}
