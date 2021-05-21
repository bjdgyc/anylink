package extension

import "testing"

func TestServerName(t *testing.T) {
	extension := ServerName{ServerName: "test.domain"}

	raw, err := extension.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	newExtension := ServerName{}
	err = newExtension.Unmarshal(raw)
	if err != nil {
		t.Fatal(err)
	}

	if newExtension.ServerName != extension.ServerName {
		t.Errorf("extensionServerName marshal: got %s expected %s", newExtension.ServerName, extension.ServerName)
	}
}
