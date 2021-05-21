package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/examples/util"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
)

func main() {
	// Prepare the IP to connect to
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4444}

	// Generate a certificate and private key to secure the connection
	certificate, genErr := selfsign.GenerateSelfSigned()
	util.Check(genErr)

	//
	// Everything below is the pion-DTLS API! Thanks for using it ❤️.
	//

	// Prepare the configuration of the DTLS connection
	config := &dtls.Config{
		Certificates:         []tls.Certificate{certificate},
		InsecureSkipVerify:   true,
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
	}

	// Connect to a DTLS server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dtlsConn, err := dtls.DialWithContext(ctx, "udp", addr, config)
	util.Check(err)
	defer func() {
		util.Check(dtlsConn.Close())
	}()

	fmt.Println("Connected; type 'exit' to shutdown gracefully")

	// Simulate a chat session
	util.Chat(dtlsConn)
}
