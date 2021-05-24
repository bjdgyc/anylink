package handshake

import (
	"reflect"
	"testing"

	"github.com/pion/dtls/v2/pkg/protocol"
)

func TestHandshakeMessageHelloVerifyRequest(t *testing.T) {
	rawHelloVerifyRequest := []byte{
		0xfe, 0xff, 0x14, 0x25, 0xfb, 0xee, 0xb3, 0x7c, 0x95, 0xcf, 0x00,
		0xeb, 0xad, 0xe2, 0xef, 0xc7, 0xfd, 0xbb, 0xed, 0xf7, 0x1f, 0x6c, 0xcd,
	}
	parsedHelloVerifyRequest := &MessageHelloVerifyRequest{
		Version: protocol.Version{Major: 0xFE, Minor: 0xFF},
		Cookie:  []byte{0x25, 0xfb, 0xee, 0xb3, 0x7c, 0x95, 0xcf, 0x00, 0xeb, 0xad, 0xe2, 0xef, 0xc7, 0xfd, 0xbb, 0xed, 0xf7, 0x1f, 0x6c, 0xcd},
	}

	h := &MessageHelloVerifyRequest{}
	if err := h.Unmarshal(rawHelloVerifyRequest); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(h, parsedHelloVerifyRequest) {
		t.Errorf("handshakeMessageClientHello unmarshal: got %#v, want %#v", h, parsedHelloVerifyRequest)
	}

	raw, err := h.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawHelloVerifyRequest) {
		t.Errorf("handshakeMessageClientHello marshal: got %#v, want %#v", raw, rawHelloVerifyRequest)
	}
}
