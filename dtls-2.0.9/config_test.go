package dtls

import (
	"crypto/dsa" //nolint
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"testing"

	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
)

func TestValidateConfig(t *testing.T) {
	// Empty config
	if err := validateConfig(nil); !errors.Is(err, errNoConfigProvided) {
		t.Fatalf("TestValidateConfig: Config validation error exp(%v) failed(%v)", errNoConfigProvided, err)
	}

	// PSK and Certificate, valid cipher suites
	cert, err := selfsign.GenerateSelfSigned()
	if err != nil {
		t.Fatalf("TestValidateConfig: Config validation error(%v), self signed certificate not generated", err)
		return
	}
	config := &Config{
		CipherSuites: []CipherSuiteID{TLS_PSK_WITH_AES_128_CCM_8, TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		PSK: func(hint []byte) ([]byte, error) {
			return nil, nil
		},
		Certificates: []tls.Certificate{cert},
	}
	if err = validateConfig(config); err != nil {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", nil, err)
	}

	// PSK and Certificate, no PSK cipher suite
	config = &Config{
		CipherSuites: []CipherSuiteID{TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		PSK: func(hint []byte) ([]byte, error) {
			return nil, nil
		},
		Certificates: []tls.Certificate{cert},
	}
	if err = validateConfig(config); !errors.Is(errNoAvailablePSKCipherSuite, err) {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", errNoAvailablePSKCipherSuite, err)
	}

	// PSK and Certificate, no non-PSK cipher suite
	config = &Config{
		CipherSuites: []CipherSuiteID{TLS_PSK_WITH_AES_128_CCM_8},
		PSK: func(hint []byte) ([]byte, error) {
			return nil, nil
		},
		Certificates: []tls.Certificate{cert},
	}
	if err = validateConfig(config); !errors.Is(errNoAvailableCertificateCipherSuite, err) {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", errNoAvailableCertificateCipherSuite, err)
	}

	// PSK identity hint with not PSK
	config = &Config{
		CipherSuites:    []CipherSuiteID{TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		PSK:             nil,
		PSKIdentityHint: []byte{},
	}
	if err = validateConfig(config); !errors.Is(err, errIdentityNoPSK) {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", errIdentityNoPSK, err)
	}

	// Invalid private key
	dsaPrivateKey := &dsa.PrivateKey{}
	err = dsa.GenerateParameters(&dsaPrivateKey.Parameters, rand.Reader, dsa.L1024N160)
	if err != nil {
		t.Fatalf("TestValidateConfig: Config validation error(%v), DSA parameters not generated", err)
		return
	}
	err = dsa.GenerateKey(dsaPrivateKey, rand.Reader)
	if err != nil {
		t.Fatalf("TestValidateConfig: Config validation error(%v), DSA private key not generated", err)
		return
	}
	config = &Config{
		CipherSuites: []CipherSuiteID{TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		Certificates: []tls.Certificate{{Certificate: cert.Certificate, PrivateKey: dsaPrivateKey}},
	}
	if err = validateConfig(config); !errors.Is(err, errInvalidPrivateKey) {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", errInvalidPrivateKey, err)
	}

	// PrivateKey without Certificate
	config = &Config{
		CipherSuites: []CipherSuiteID{TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		Certificates: []tls.Certificate{{PrivateKey: cert.PrivateKey}},
	}
	if err = validateConfig(config); !errors.Is(err, errInvalidCertificate) {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", errInvalidCertificate, err)
	}

	// Invalid cipher suites
	config = &Config{CipherSuites: []CipherSuiteID{0x0000}}
	if err = validateConfig(config); err == nil {
		t.Fatal("TestValidateConfig: Client error expected with invalid CipherSuiteID")
	}

	// Valid config
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("TestValidateConfig: Config validation error(%v), RSA private key not generated", err)
		return
	}
	config = &Config{
		CipherSuites: []CipherSuiteID{TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		Certificates: []tls.Certificate{cert, {Certificate: cert.Certificate, PrivateKey: rsaPrivateKey}},
	}
	if err = validateConfig(config); err != nil {
		t.Fatalf("TestValidateConfig: Client error exp(%v) failed(%v)", nil, err)
	}
}
