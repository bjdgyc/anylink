// Package util provides auxiliary utilities used in examples
package util

import (
	"bufio"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const bufSize = 8192

var (
	errBlockIsNotPrivateKey  = errors.New("block is not a private key, unable to load key")
	errUnknownKeyTime        = errors.New("unknown key time in PKCS#8 wrapping, unable to load key")
	errNoPrivateKeyFound     = errors.New("no private key found, unable to load key")
	errBlockIsNotCertificate = errors.New("block is not a certificate, unable to load certificates")
	errNoCertificateFound    = errors.New("no certificate found, unable to load certificates")
)

// Chat simulates a simple text chat session over the connection
func Chat(conn io.ReadWriter) {
	go func() {
		b := make([]byte, bufSize)

		for {
			n, err := conn.Read(b)
			Check(err)
			fmt.Printf("Got message: %s\n", string(b[:n]))
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		Check(err)

		if strings.TrimSpace(text) == "exit" {
			return
		}

		_, err = conn.Write([]byte(text))
		Check(err)
	}
}

// Check is a helper to throw errors in the examples
func Check(err error) {
	switch e := err.(type) {
	case nil:
	case (net.Error):
		if e.Temporary() {
			fmt.Printf("Warning: %v\n", err)
			return
		}

		fmt.Printf("net.Error: %v\n", err)
		panic(err)
	default:
		fmt.Printf("error: %v\n", err)
		panic(err)
	}
}

// LoadKeyAndCertificate reads certificates or key from file
func LoadKeyAndCertificate(keyPath string, certificatePath string) (*tls.Certificate, error) {
	privateKey, err := LoadKey(keyPath)
	if err != nil {
		return nil, err
	}

	certificate, err := LoadCertificate(certificatePath)
	if err != nil {
		return nil, err
	}

	certificate.PrivateKey = privateKey

	return certificate, nil
}

// LoadKey Load/read key from file
func LoadKey(path string) (crypto.PrivateKey, error) {
	rawData, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(rawData)
	if block == nil || !strings.HasSuffix(block.Type, "PRIVATE KEY") {
		return nil, errBlockIsNotPrivateKey
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errUnknownKeyTime
		}
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, errNoPrivateKeyFound
}

// LoadCertificate Load/read certificate(s) from file
func LoadCertificate(path string) (*tls.Certificate, error) {
	rawData, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var certificate tls.Certificate

	for {
		block, rest := pem.Decode(rawData)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			return nil, errBlockIsNotCertificate
		}

		certificate.Certificate = append(certificate.Certificate, block.Bytes)
		rawData = rest
	}

	if len(certificate.Certificate) == 0 {
		return nil, errNoCertificateFound
	}

	return &certificate, nil
}
