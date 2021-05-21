package extension

import (
	"errors"
	"testing"
)

func TestExtensions(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		extensions, err := Unmarshal([]byte{})
		if err != nil || len(extensions) != 0 {
			t.Fatal("Failed to decode zero extensions")
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		extensions, err := Unmarshal([]byte{0x00})
		if !errors.Is(err, errBufferTooSmall) || len(extensions) != 0 {
			t.Fatal("Failed to error on invalid extension")
		}
	})
}
