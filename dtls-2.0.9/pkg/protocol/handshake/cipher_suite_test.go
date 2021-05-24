package handshake

import (
	"errors"
	"testing"
)

func TestDecodeCipherSuiteIDs(t *testing.T) {
	testCases := []struct {
		buf    []byte
		result []uint16
		err    error
	}{
		{[]byte{}, nil, errBufferTooSmall},
	}

	for _, testCase := range testCases {
		_, err := decodeCipherSuiteIDs(testCase.buf)
		if !errors.Is(err, testCase.err) {
			t.Fatal("Unexpected error", err)
		}
	}
}
