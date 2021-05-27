package handshake

import (
	"reflect"
	"testing"
)

func TestHandshakeMessageFinished(t *testing.T) {
	rawFinished := []byte{
		0x01, 0x01, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	}
	parsedFinished := &MessageFinished{
		VerifyData: rawFinished,
	}

	c := &MessageFinished{}
	if err := c.Unmarshal(rawFinished); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(c, parsedFinished) {
		t.Errorf("handshakeMessageFinished unmarshal: got %#v, want %#v", c, parsedFinished)
	}

	raw, err := c.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawFinished) {
		t.Errorf("handshakeMessageFinished marshal: got %#v, want %#v", raw, rawFinished)
	}
}
