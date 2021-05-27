package protocol

import (
	"errors"
	"testing"
)

func TestDecodeCompressionMethods(t *testing.T) {
	testCases := []struct {
		buf    []byte
		result []*CompressionMethod
		err    error
	}{
		{[]byte{}, nil, errBufferTooSmall},
	}

	for _, testCase := range testCases {
		_, err := DecodeCompressionMethods(testCase.buf)
		if !errors.Is(err, testCase.err) {
			t.Fatal("Unexpected error", err)
		}
	}
}
