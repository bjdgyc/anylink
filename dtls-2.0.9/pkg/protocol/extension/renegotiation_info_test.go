package extension

import "testing"

func TestRenegotiationInfo(t *testing.T) {
	extension := RenegotiationInfo{RenegotiatedConnection: 0}

	raw, err := extension.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	newExtension := RenegotiationInfo{}
	err = newExtension.Unmarshal(raw)
	if err != nil {
		t.Fatal(err)
	}

	if newExtension.RenegotiatedConnection != extension.RenegotiatedConnection {
		t.Errorf("extensionRenegotiationInfo marshal: got %d expected %d", newExtension.RenegotiatedConnection, extension.RenegotiatedConnection)
	}
}
