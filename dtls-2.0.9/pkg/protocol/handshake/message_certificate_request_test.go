package handshake

import (
	"reflect"
	"testing"

	"github.com/pion/dtls/v2/pkg/crypto/clientcertificate"
	"github.com/pion/dtls/v2/pkg/crypto/hash"
	"github.com/pion/dtls/v2/pkg/crypto/signature"
	"github.com/pion/dtls/v2/pkg/crypto/signaturehash"
)

func TestHandshakeMessageCertificateRequest(t *testing.T) {
	rawCertificateRequest := []byte{
		0x02, 0x01, 0x40, 0x00, 0x0C, 0x04, 0x03, 0x04, 0x01, 0x05,
		0x03, 0x05, 0x01, 0x06, 0x01, 0x02, 0x01, 0x00, 0x00,
	}
	parsedCertificateRequest := &MessageCertificateRequest{
		CertificateTypes: []clientcertificate.Type{
			clientcertificate.RSASign,
			clientcertificate.ECDSASign,
		},
		SignatureHashAlgorithms: []signaturehash.Algorithm{
			{Hash: hash.SHA256, Signature: signature.ECDSA},
			{Hash: hash.SHA256, Signature: signature.RSA},
			{Hash: hash.SHA384, Signature: signature.ECDSA},
			{Hash: hash.SHA384, Signature: signature.RSA},
			{Hash: hash.SHA512, Signature: signature.RSA},
			{Hash: hash.SHA1, Signature: signature.RSA},
		},
	}

	c := &MessageCertificateRequest{}
	if err := c.Unmarshal(rawCertificateRequest); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(c, parsedCertificateRequest) {
		t.Errorf("parsedCertificateRequest unmarshal: got %#v, want %#v", c, parsedCertificateRequest)
	}

	raw, err := c.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawCertificateRequest) {
		t.Errorf("parsedCertificateRequest marshal: got %#v, want %#v", raw, rawCertificateRequest)
	}
}
