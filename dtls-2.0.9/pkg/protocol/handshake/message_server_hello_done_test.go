package handshake

import (
	"reflect"
	"testing"
)

func TestHandshakeMessageServerHelloDone(t *testing.T) {
	rawServerHelloDone := []byte{}
	parsedServerHelloDone := &MessageServerHelloDone{}

	c := &MessageServerHelloDone{}
	if err := c.Unmarshal(rawServerHelloDone); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(c, parsedServerHelloDone) {
		t.Errorf("handshakeMessageServerHelloDone unmarshal: got %#v, want %#v", c, parsedServerHelloDone)
	}

	raw, err := c.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawServerHelloDone) {
		t.Errorf("handshakeMessageServerHelloDone marshal: got %#v, want %#v", raw, rawServerHelloDone)
	}
}
