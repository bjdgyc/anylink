package sessdata

import (
	"fmt"
	"testing"

	"github.com/bjdgyc/anylink/base"
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

	var (
		cmpName string
		ok      bool
	)
	base.Cfg.Compression = true

	cmpName, ok = cSess.SetPickCmp("cstp", "oc-lz4,lzs")
	fmt.Println(cmpName, ok)
	ast.True(ok)
	ast.Equal(cmpName, "lzs")
	cmpName, ok = cSess.SetPickCmp("dtls", "lzs")
	ast.True(ok)
	ast.Equal(cmpName, "lzs")
	cmpName, ok = cSess.SetPickCmp("dtls", "test")
	ast.False(ok)
	ast.Equal(cmpName, "")

	cSess.Close()
}
