package handler

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"github.com/pion/logging"
)

// 因本项目对 github.com/pion/dtls 的代码，进行了大量的修改
// 且短时间内无法合并到上游项目
// 所以本项目暂时copy了一份代码
// 最后,感谢 github.com/pion/dtls 对golang生态做出的贡献

func startDtls() {
	certificate, err := selfsign.GenerateSelfSigned()
	if err != nil {
		panic(err)
	}
	logf := logging.NewDefaultLoggerFactory()
	logf.Writer = base.GetBaseLw()
	// logf.DefaultLogLevel = logging.LogLevelTrace
	logf.DefaultLogLevel = logging.LogLevelInfo

	config := &dtls.Config{
		Certificates:         []tls.Certificate{certificate},
		InsecureSkipVerify:   true,
		ExtendedMasterSecret: dtls.DisableExtendedMasterSecret,
		CipherSuites:         []dtls.CipherSuiteID{dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		LoggerFactory:        logf,
		MTU:                  BufferSize,
		CiscoCompat: func(sessid []byte) ([]byte, error) {
			masterSecret := sessdata.Dtls2MasterSecret(hex.EncodeToString(sessid))
			if masterSecret == "" {
				return nil, fmt.Errorf("masterSecret is err")
			}
			return hex.DecodeString(masterSecret)
		},
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
