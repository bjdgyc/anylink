package admin

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespSucess(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	RespSucess(w, "data")
	// fmt.Println(w)
	assert.Equal(w.Code, 200)
	body, _ := ioutil.ReadAll(w.Body)
	res := Resp{}
	err := json.Unmarshal(body, &res)
	assert.Nil(err)
	assert.Equal(res.Code, 0)
	assert.Equal(res.Data, "data")

}

func TestRespError(t *testing.T) {
	assert := assert.New(t)
	w := httptest.NewRecorder()
	RespError(w, 10, "err-msg")
	// fmt.Println(w)
	assert.Equal(w.Code, 200)
	body, _ := ioutil.ReadAll(w.Body)
	res := Resp{}
	err := json.Unmarshal(body, &res)
	assert.Nil(err)
	assert.Equal(res.Code, 10)
	assert.Equal(res.Msg, "err-msg")
}
