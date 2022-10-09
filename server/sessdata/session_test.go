package sessdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	ast := assert.New(t)
	sessions = make(map[string]*Session)
	sess := NewSession("")
	token := sess.Token
	v, ok := sessions[token]
	ast.True(ok)
	ast.Equal(sess, v)
}

func TestConnSession(t *testing.T) {
	ast := assert.New(t)
	tmp := t.TempDir()
	preData(tmp)
	defer cleardata(tmp)

	sess := NewSession("")
	sess.Group = "group1"
	sess.MacAddr = "00:15:5d:50:14:43"

	cSess := sess.NewConn()

	err := cSess.RateLimit(100, true)
	ast.Nil(err)
	ast.Equal(cSess.BandwidthUp.Load(), uint32(100))
	err = cSess.RateLimit(200, false)
	ast.Nil(err)
	ast.Equal(cSess.BandwidthDown.Load(), uint32(200))
	cSess.Close()
}
