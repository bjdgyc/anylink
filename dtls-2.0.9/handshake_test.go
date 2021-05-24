package dtls

import (
	"reflect"
	"testing"
	"time"

	"github.com/pion/dtls/v2/pkg/protocol"
	"github.com/pion/dtls/v2/pkg/protocol/extension"
	"github.com/pion/dtls/v2/pkg/protocol/handshake"
)

func TestHandshakeMessage(t *testing.T) {
	rawHandshakeMessage := []byte{
		0x01, 0x00, 0x00, 0x29, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x29, 0xfe, 0xfd, 0xb6,
		0x2f, 0xce, 0x5c, 0x42, 0x54, 0xff, 0x86, 0xe1, 0x24, 0x41, 0x91, 0x42, 0x62, 0x15, 0xad,
		0x16, 0xc9, 0x15, 0x8d, 0x95, 0x71, 0x8a, 0xbb, 0x22, 0xd7, 0x47, 0xec, 0xd8, 0x3d, 0xdc,
		0x4b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	parsedHandshake := &handshake.Handshake{
		Header: handshake.Header{
			Length:         0x29,
			FragmentLength: 0x29,
			Type:           handshake.TypeClientHello,
		},
		Message: &handshake.MessageClientHello{
			Version: protocol.Version{Major: 0xFE, Minor: 0xFD},
			Random: handshake.Random{
				GMTUnixTime: time.Unix(3056586332, 0),
				RandomBytes: [28]byte{0x42, 0x54, 0xff, 0x86, 0xe1, 0x24, 0x41, 0x91, 0x42, 0x62, 0x15, 0xad, 0x16, 0xc9, 0x15, 0x8d, 0x95, 0x71, 0x8a, 0xbb, 0x22, 0xd7, 0x47, 0xec, 0xd8, 0x3d, 0xdc, 0x4b},
			},
			Cookie:             []byte{},
			CipherSuiteIDs:     []uint16{},
			CompressionMethods: []*protocol.CompressionMethod{},
			Extensions:         []extension.Extension{},
		},
	}

	h := &handshake.Handshake{}
	if err := h.Unmarshal(rawHandshakeMessage); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(h, parsedHandshake) {
		t.Errorf("handshakeMessageClientHello unmarshal: got %#v, want %#v", h, parsedHandshake)
	}

	raw, err := h.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawHandshakeMessage) {
		t.Errorf("handshakeMessageClientHello marshal: got %#v, want %#v", raw, rawHandshakeMessage)
	}
}
