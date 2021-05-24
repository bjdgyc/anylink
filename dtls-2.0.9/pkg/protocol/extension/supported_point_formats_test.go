package extension

import (
	"reflect"
	"testing"

	"github.com/pion/dtls/v2/pkg/crypto/elliptic"
)

func TestExtensionSupportedPointFormats(t *testing.T) {
	rawExtensionSupportedPointFormats := []byte{0x00, 0x0b, 0x00, 0x02, 0x01, 0x00}
	parsedExtensionSupportedPointFormats := &SupportedPointFormats{
		PointFormats: []elliptic.CurvePointFormat{elliptic.CurvePointFormatUncompressed},
	}

	raw, err := parsedExtensionSupportedPointFormats.Marshal()
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(raw, rawExtensionSupportedPointFormats) {
		t.Errorf("extensionSupportedPointFormats marshal: got %#v, want %#v", raw, rawExtensionSupportedPointFormats)
	}
}
