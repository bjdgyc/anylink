package handshake

import (
	"reflect"
	"testing"
)

func TestHandshakeMessageClientKeyExchange(t *testing.T) {
	rawClientKeyExchange := []byte{
		0x20, 0x26, 0x78, 0x4a, 0x78, 0x70, 0xc1, 0xf9, 0x71, 0xea, 0x50, 0x4a, 0xb5, 0xbb, 0x00, 0x76,
		0x02, 0x05, 0xda, 0xf7, 0xd0, 0x3f, 0xe3, 0xf7, 0x4e, 0x8a, 0x14, 0x6f, 0xb7, 0xe0, 0xc0, 0xff,
		0x54,
	}
	parsedClientKeyExchange := &MessageClientKeyExchange{
		PublicKey: rawClientKeyExchange[1:],
	}

	c := &MessageClientKeyExchange{}
	if err := c.Unmarshal(rawClientKeyExchange); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(c, parsedClientKeyExchange) {
		t.Errorf("handshakeMessageClientKeyExchange unmarshal: got %#v, want %#v", c, parsedClientKeyExchange)
	}

	raw, err := c.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawClientKeyExchange) {
		t.Errorf("handshakeMessageClientKeyExchange marshal: got %#v, want %#v", raw, rawClientKeyExchange)
	}
}
