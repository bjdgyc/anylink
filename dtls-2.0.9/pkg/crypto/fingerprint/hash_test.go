package fingerprint

import (
	"crypto"
	"errors"
	"testing"
)

func TestHashFromString(t *testing.T) {
	t.Run("InvalidHashAlgorithm", func(t *testing.T) {
		_, err := HashFromString("invalid-hash-algorithm")
		if !errors.Is(err, errInvalidHashAlgorithm) {
			t.Errorf("Expected error '%v' for invalid hash name, got '%v'", errInvalidHashAlgorithm, err)
		}
	})
	t.Run("ValidHashAlgorithm", func(t *testing.T) {
		h, err := HashFromString("sha-512")
		if err != nil {
			t.Fatalf("Unexpected error for valid hash name, got '%v'", err)
		}
		if h != crypto.SHA512 {
			t.Errorf("Expected hash ID of %d, got %d", int(crypto.SHA512), int(h))
		}
	})
}

func TestStringFromHash_Roundtrip(t *testing.T) {
	for _, h := range nameToHash() {
		s, err := StringFromHash(h)
		if err != nil {
			t.Fatalf("Unexpected error for valid hash algorithm, got '%v'", err)
		}
		h2, err := HashFromString(s)
		if err != nil {
			t.Fatalf("Unexpected error for valid hash name, got '%v'", err)
		}
		if h != h2 {
			t.Errorf("Hash value doesn't match, expected: 0x%x, got 0x%x", h, h2)
		}
	}
}
