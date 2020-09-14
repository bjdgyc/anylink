package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInArrStr(t *testing.T) {
	assert := assert.New(t)
	arr := []string{"a", "b", "c"}
	assert.True(InArrStr(arr, "b"))
	assert.False(InArrStr(arr, "d"))
}

func TestHumanByte(t *testing.T) {
	assert := assert.New(t)
	var s string
	s = HumanByte(999)
	assert.Equal(s, "999.00 B")
	s = HumanByte(10256)
	assert.Equal(s, "10.02 KB")
	s = HumanByte(99 * 1024 * 1024)
	assert.Equal(s, "99.00 MB")
	s = HumanByte(1023 * 1024 * 1024)
	assert.Equal(s, "1023.00 MB")
	s = HumanByte(1024 * 1024 * 1024)
	assert.Equal(s, "1.00 GB")
}
