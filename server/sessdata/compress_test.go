package sessdata

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLzsCompress(t *testing.T) {
	var (
		n   int
		err error
	)
	assert := assert.New(t)
	c := LzsgoCmp{}
	s := "hello anylink, you are best!"
	src := []byte(strings.Repeat(s, 50))

	comprBuf := make([]byte, 2048)
	n, err = c.Compress(src, comprBuf)
	assert.Nil(err)

	unprBuf := make([]byte, 2048)
	n, err = c.Uncompress(comprBuf[:n], unprBuf)
	assert.Nil(err)
	assert.Equal(src, unprBuf[:n])
}
