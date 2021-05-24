// +build openssl,go1.13,!js

package e2e

import (
	"testing"
)

func TestPionOpenSSLE2ESimpleED25519(t *testing.T) {
	t.Skip("TODO: waiting OpenSSL's DTLS Ed25519 support")
	t.Run("OpenSSLServer", func(t *testing.T) {
		testPionE2ESimpleED25519(t, serverOpenSSL, clientPion)
	})
	t.Run("OpenSSLClient", func(t *testing.T) {
		testPionE2ESimpleED25519(t, serverPion, clientOpenSSL)
	})
}
