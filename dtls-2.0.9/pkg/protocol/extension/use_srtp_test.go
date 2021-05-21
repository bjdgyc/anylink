package extension

import (
	"reflect"
	"testing"
)

func TestExtensionUseSRTP(t *testing.T) {
	rawUseSRTP := []byte{0x00, 0x0e, 0x00, 0x05, 0x00, 0x02, 0x00, 0x01, 0x00}
	parsedUseSRTP := &UseSRTP{
		ProtectionProfiles: []SRTPProtectionProfile{SRTP_AES128_CM_HMAC_SHA1_80},
	}

	raw, err := parsedUseSRTP.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawUseSRTP) {
		t.Errorf("extensionUseSRTP marshal: got %#v, want %#v", raw, rawUseSRTP)
	}
}
