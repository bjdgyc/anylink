package handler

import (
	"crypto/tls"
	"encoding/hex"
	"log"
	"net"
	"time"
	"os"

	"github.com/bjdgyc/anylink/sessdata"
	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/logging"
)

func startDtls() {
	certificate, err := selfsign.GenerateSelfSigned()

	logf := logging.NewDefaultLoggerFactory()
	logf.DefaultLogLevel = logging.LogLevelTrace
	f, err := os.OpenFile("/tmp/key.log", os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	config := &dtls.Config{
		Certificates:         []tls.Certificate{certificate},
		InsecureSkipVerify:   true,
		ExtendedMasterSecret: dtls.DisableExtendedMasterSecret,
		CipherSuites:         []dtls.CipherSuiteID{dtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		LoggerFactory:        logf,
		KeyLogWriter:         f,
	}

	addr := &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 4433}

	ln, err := dtls.Listen("udp", addr, config)
	if err != nil {
		panic(err)
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("Accept error", err)
			continue
		}

		go func() {
			time.Sleep(1 * time.Second)
			cc := c.(*dtls.Conn)
			id := hex.EncodeToString(cc.ConnectionState().SessionID)
			s, ok := ss.Load(id)
			log.Println("get link", id, ok)
			cs := s.(*sessdata.ConnSession)
			LinkDtls(c, cs)
		}()
	}
}
