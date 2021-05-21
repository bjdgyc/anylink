package protocol

import (
	"errors"
	"reflect"
	"testing"
)

func TestChangeCipherSpecRoundTrip(t *testing.T) {
	c := ChangeCipherSpec{}
	raw, err := c.Marshal()
	if err != nil {
		t.Error(err)
	}

	var cNew ChangeCipherSpec
	if err := cNew.Unmarshal(raw); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(c, cNew) {
		t.Errorf("ChangeCipherSpec round trip: got %#v, want %#v", cNew, c)
	}
}

func TestChangeCipherSpecInvalid(t *testing.T) {
	c := ChangeCipherSpec{}
	if err := c.Unmarshal([]byte{0x00}); !errors.Is(err, errInvalidCipherSpec) {
		t.Errorf("ChangeCipherSpec invalid assert: got %#v, want %#v", err, errInvalidCipherSpec)
	}
}
