// +build go1.14

package dtls

import (
	"testing"
)

func TestInsecureCipherSuites(t *testing.T) {
	r := InsecureCipherSuites()

	if len(r) != 0 {
		t.Fatalf("Expected no insecure ciphersuites, got %d", len(r))
	}
}

func TestCipherSuites(t *testing.T) {
	ours := allCipherSuites()
	theirs := CipherSuites()

	if len(ours) != len(theirs) {
		t.Fatalf("Expected %d CipherSuites, got %d", len(ours), len(theirs))
	}

	for i, s := range ours {
		i := i
		s := s
		t.Run(s.String(), func(t *testing.T) {
			c := theirs[i]
			if c.ID != uint16(s.ID()) {
				t.Fatalf("Expected ID: 0x%04X, got 0x%04X", s.ID(), c.ID)
			}

			if c.Name != s.String() {
				t.Fatalf("Expected Name: %s, got %s", s.String(), c.Name)
			}

			if len(c.SupportedVersions) != 1 {
				t.Fatalf("Expected %d SupportedVersion, got %d", 1, len(c.SupportedVersions))
			}

			if c.SupportedVersions[0] != VersionDTLS12 {
				t.Fatalf("Expected SupportedVersions 0x%04X, got 0x%04X", VersionDTLS12, c.SupportedVersions[0])
			}

			if c.Insecure {
				t.Fatalf("Expected Insecure %t, got %t", false, c.Insecure)
			}
		})
	}
}
