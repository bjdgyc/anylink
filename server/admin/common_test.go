package admin

import (
	"testing"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/stretchr/testify/assert"
)

func TestJwtData(t *testing.T) {
	assert := assert.New(t)
	base.Cfg.JwtSecret = "dsfasfdfsadfasdfasd3sdaf"
	data := map[string]interface{}{
		"key": "value",
	}
	expiresAt := time.Now().Add(time.Minute).Unix()
	token, err := SetJwtData(data, expiresAt)
	assert.Nil(err)
	dataN, err := GetJwtData(token)
	assert.Nil(err)
	assert.Equal(dataN["key"], "value")
}
