package alert

import (
	"errors"
	"reflect"
	"testing"
)

func TestAlert(t *testing.T) {
	for _, test := range []struct {
		Name               string
		Data               []byte
		Want               *Alert
		WantUnmarshalError error
	}{
		{
			Name: "Valid Alert",
			Data: []byte{0x02, 0x0A},
			Want: &Alert{
				Level:       Fatal,
				Description: UnexpectedMessage,
			},
		},
		{
			Name:               "Invalid alert length",
			Data:               []byte{0x00},
			Want:               &Alert{},
			WantUnmarshalError: errBufferTooSmall,
		},
	} {
		a := &Alert{}
		if err := a.Unmarshal(test.Data); !errors.Is(err, test.WantUnmarshalError) {
			t.Errorf("Unexpected Error %v: exp: %v got: %v", test.Name, test.WantUnmarshalError, err)
		} else if !reflect.DeepEqual(test.Want, a) {
			t.Errorf("%q alert.unmarshal: got %v, want %v", test.Name, a, test.Want)
		}

		if test.WantUnmarshalError != nil {
			return
		}

		data, marshalErr := a.Marshal()
		if marshalErr != nil {
			t.Errorf("Unexpected Error %v: got: %v", test.Name, marshalErr)
		} else if !reflect.DeepEqual(test.Data, data) {
			t.Errorf("%q alert.marshal: got % 02x, want % 02x", test.Name, data, test.Data)
		}
	}
}
