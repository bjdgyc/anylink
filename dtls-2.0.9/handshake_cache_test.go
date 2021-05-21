package dtls

import (
	"bytes"
	"testing"

	"github.com/pion/dtls/v2/internal/ciphersuite"
	"github.com/pion/dtls/v2/pkg/protocol/handshake"
)

func TestHandshakeCacheSinglePush(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Rule     []handshakeCachePullRule
		Input    []handshakeCacheItem
		Expected []byte
	}{
		{
			Name: "Single Push",
			Input: []handshakeCacheItem{
				{0, true, 0, 0, []byte{0x00}},
			},
			Rule: []handshakeCachePullRule{
				{0, 0, true, false},
			},
			Expected: []byte{0x00},
		},
		{
			Name: "Multi Push",
			Input: []handshakeCacheItem{
				{0, true, 0, 0, []byte{0x00}},
				{1, true, 0, 1, []byte{0x01}},
				{2, true, 0, 2, []byte{0x02}},
			},
			Rule: []handshakeCachePullRule{
				{0, 0, true, false},
				{1, 0, true, false},
				{2, 0, true, false},
			},
			Expected: []byte{0x00, 0x01, 0x02},
		},
		{
			Name: "Multi Push, Rules set order",
			Input: []handshakeCacheItem{
				{2, true, 0, 2, []byte{0x02}},
				{0, true, 0, 0, []byte{0x00}},
				{1, true, 0, 1, []byte{0x01}},
			},
			Rule: []handshakeCachePullRule{
				{0, 0, true, false},
				{1, 0, true, false},
				{2, 0, true, false},
			},
			Expected: []byte{0x00, 0x01, 0x02},
		},

		{
			Name: "Multi Push, Dupe Seqnum",
			Input: []handshakeCacheItem{
				{0, true, 0, 0, []byte{0x00}},
				{1, true, 0, 1, []byte{0x01}},
				{1, true, 0, 1, []byte{0x01}},
			},
			Rule: []handshakeCachePullRule{
				{0, 0, true, false},
				{1, 0, true, false},
			},
			Expected: []byte{0x00, 0x01},
		},
		{
			Name: "Multi Push, Dupe Seqnum Client/Server",
			Input: []handshakeCacheItem{
				{0, true, 0, 0, []byte{0x00}},
				{1, true, 0, 1, []byte{0x01}},
				{1, false, 0, 1, []byte{0x02}},
			},
			Rule: []handshakeCachePullRule{
				{0, 0, true, false},
				{1, 0, true, false},
				{1, 0, false, false},
			},
			Expected: []byte{0x00, 0x01, 0x02},
		},
		{
			Name: "Multi Push, Dupe Seqnum with Unique HandshakeType",
			Input: []handshakeCacheItem{
				{1, true, 0, 0, []byte{0x00}},
				{2, true, 0, 1, []byte{0x01}},
				{3, false, 0, 0, []byte{0x02}},
			},
			Rule: []handshakeCachePullRule{
				{1, 0, true, false},
				{2, 0, true, false},
				{3, 0, false, false},
			},
			Expected: []byte{0x00, 0x01, 0x02},
		},
		{
			Name: "Multi Push, Wrong epoch",
			Input: []handshakeCacheItem{
				{1, true, 0, 0, []byte{0x00}},
				{2, true, 1, 1, []byte{0x01}},
				{2, true, 0, 2, []byte{0x11}},
				{3, false, 0, 0, []byte{0x02}},
				{3, false, 1, 0, []byte{0x12}},
				{3, false, 2, 0, []byte{0x12}},
			},
			Rule: []handshakeCachePullRule{
				{1, 0, true, false},
				{2, 1, true, false},
				{3, 0, false, false},
			},
			Expected: []byte{0x00, 0x01, 0x02},
		},
	} {
		h := newHandshakeCache()
		for _, i := range test.Input {
			h.push(i.data, i.epoch, i.messageSequence, i.typ, i.isClient)
		}
		verifyData := h.pullAndMerge(test.Rule...)
		if !bytes.Equal(verifyData, test.Expected) {
			t.Errorf("handshakeCache '%s' exp: % 02x actual % 02x", test.Name, test.Expected, verifyData)
		}
	}
}

func TestHandshakeCacheSessionHash(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Rule     []handshakeCachePullRule
		Input    []handshakeCacheItem
		Expected []byte
	}{
		{
			Name: "Standard Handshake",
			Input: []handshakeCacheItem{
				{handshake.TypeClientHello, true, 0, 0, []byte{0x00}},
				{handshake.TypeServerHello, false, 0, 1, []byte{0x01}},
				{handshake.TypeCertificate, false, 0, 2, []byte{0x02}},
				{handshake.TypeServerKeyExchange, false, 0, 3, []byte{0x03}},
				{handshake.TypeServerHelloDone, false, 0, 4, []byte{0x04}},
				{handshake.TypeClientKeyExchange, true, 0, 5, []byte{0x05}},
			},
			Expected: []byte{0x17, 0xe8, 0x8d, 0xb1, 0x87, 0xaf, 0xd6, 0x2c, 0x16, 0xe5, 0xde, 0xbf, 0x3e, 0x65, 0x27, 0xcd, 0x00, 0x6b, 0xc0, 0x12, 0xbc, 0x90, 0xb5, 0x1a, 0x81, 0x0c, 0xd8, 0x0c, 0x2d, 0x51, 0x1f, 0x43},
		},
		{
			Name: "Handshake With Client Cert Request",
			Input: []handshakeCacheItem{
				{handshake.TypeClientHello, true, 0, 0, []byte{0x00}},
				{handshake.TypeServerHello, false, 0, 1, []byte{0x01}},
				{handshake.TypeCertificate, false, 0, 2, []byte{0x02}},
				{handshake.TypeServerKeyExchange, false, 0, 3, []byte{0x03}},
				{handshake.TypeCertificateRequest, false, 0, 4, []byte{0x04}},
				{handshake.TypeServerHelloDone, false, 0, 5, []byte{0x05}},
				{handshake.TypeClientKeyExchange, true, 0, 6, []byte{0x06}},
			},
			Expected: []byte{0x57, 0x35, 0x5a, 0xc3, 0x30, 0x3c, 0x14, 0x8f, 0x11, 0xae, 0xf7, 0xcb, 0x17, 0x94, 0x56, 0xb9, 0x23, 0x2c, 0xde, 0x33, 0xa8, 0x18, 0xdf, 0xda, 0x2c, 0x2f, 0xcb, 0x93, 0x25, 0x74, 0x9a, 0x6b},
		},
		{
			Name: "Handshake Ignores after ClientKeyExchange",
			Input: []handshakeCacheItem{
				{handshake.TypeClientHello, true, 0, 0, []byte{0x00}},
				{handshake.TypeServerHello, false, 0, 1, []byte{0x01}},
				{handshake.TypeCertificate, false, 0, 2, []byte{0x02}},
				{handshake.TypeServerKeyExchange, false, 0, 3, []byte{0x03}},
				{handshake.TypeCertificateRequest, false, 0, 4, []byte{0x04}},
				{handshake.TypeServerHelloDone, false, 0, 5, []byte{0x05}},
				{handshake.TypeClientKeyExchange, true, 0, 6, []byte{0x06}},
				{handshake.TypeCertificateVerify, true, 0, 7, []byte{0x07}},
				{handshake.TypeFinished, true, 1, 7, []byte{0x08}},
				{handshake.TypeFinished, false, 1, 7, []byte{0x09}},
			},
			Expected: []byte{0x57, 0x35, 0x5a, 0xc3, 0x30, 0x3c, 0x14, 0x8f, 0x11, 0xae, 0xf7, 0xcb, 0x17, 0x94, 0x56, 0xb9, 0x23, 0x2c, 0xde, 0x33, 0xa8, 0x18, 0xdf, 0xda, 0x2c, 0x2f, 0xcb, 0x93, 0x25, 0x74, 0x9a, 0x6b},
		},
		{
			Name: "Handshake Ignores wrong epoch",
			Input: []handshakeCacheItem{
				{handshake.TypeClientHello, true, 0, 0, []byte{0x00}},
				{handshake.TypeServerHello, false, 0, 1, []byte{0x01}},
				{handshake.TypeCertificate, false, 0, 2, []byte{0x02}},
				{handshake.TypeServerKeyExchange, false, 0, 3, []byte{0x03}},
				{handshake.TypeCertificateRequest, false, 0, 4, []byte{0x04}},
				{handshake.TypeServerHelloDone, false, 0, 5, []byte{0x05}},
				{handshake.TypeClientKeyExchange, true, 0, 6, []byte{0x06}},
				{handshake.TypeCertificateVerify, true, 0, 7, []byte{0x07}},
				{handshake.TypeFinished, true, 0, 7, []byte{0xf0}},
				{handshake.TypeFinished, false, 0, 7, []byte{0xf1}},
				{handshake.TypeFinished, true, 1, 7, []byte{0x08}},
				{handshake.TypeFinished, false, 1, 7, []byte{0x09}},
				{handshake.TypeFinished, true, 0, 7, []byte{0xf0}},
				{handshake.TypeFinished, false, 0, 7, []byte{0xf1}},
			},
			Expected: []byte{0x57, 0x35, 0x5a, 0xc3, 0x30, 0x3c, 0x14, 0x8f, 0x11, 0xae, 0xf7, 0xcb, 0x17, 0x94, 0x56, 0xb9, 0x23, 0x2c, 0xde, 0x33, 0xa8, 0x18, 0xdf, 0xda, 0x2c, 0x2f, 0xcb, 0x93, 0x25, 0x74, 0x9a, 0x6b},
		},
	} {
		h := newHandshakeCache()
		for _, i := range test.Input {
			h.push(i.data, i.epoch, i.messageSequence, i.typ, i.isClient)
		}

		cipherSuite := ciphersuite.TLSEcdheEcdsaWithAes128GcmSha256{}
		verifyData, err := h.sessionHash(cipherSuite.HashFunc(), 0)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(verifyData, test.Expected) {
			t.Errorf("handshakeCacheSesssionHassh '%s' exp: % 02x actual % 02x", test.Name, test.Expected, verifyData)
		}
	}
}
