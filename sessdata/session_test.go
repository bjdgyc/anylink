package sessdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	assert := assert.New(t)
	sessions = make(map[string]*Session)
	sess := NewSession("")
	token := sess.Token
	v, ok := sessions[token]
	assert.True(ok)
	assert.Equal(sess, v)
}

func TestConnSession(t *testing.T) {
	assert := assert.New(t)
	preIpData()
	defer closeIpdata()
	sess := NewSession("")
	cSess := sess.NewConn()
	cSess.RateLimit(100, true)
	assert.Equal(cSess.BandwidthUp, uint32(100))
	cSess.RateLimit(200, false)
	assert.Equal(cSess.BandwidthDown, uint32(200))
	cSess.Close()
}
