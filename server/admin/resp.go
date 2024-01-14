package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/bjdgyc/anylink/base"
)

type Resp struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Location string `json:"location"`
	Data     any    `json:"data"`
}

func respHttp(w http.ResponseWriter, respCode int, data any, errS ...any) {
	resp := Resp{
		Code: respCode,
		Msg:  "success",
		Data: data,
	}
	_, file, line, _ := runtime.Caller(2)
	resp.Location = fmt.Sprintf("%v:%v", file, line)

	if respCode != 0 {
		resp.Msg = ""
		if v, ok := RespMap[respCode]; ok {
			resp.Msg += v
		}

		if len(errS) > 0 {
			resp.Msg += fmt.Sprint(errS...)
		}
	}

	b, err := json.Marshal(resp)
	if err != nil {
		base.Error(err, resp)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		base.Error(err)
	}
	// 记录返回数据
	// logger.Category("response").Debug(string(b))
}

func RespSucess(w http.ResponseWriter, data any) {
	respHttp(w, 0, data, "")
}

func RespError(w http.ResponseWriter, respCode int, errS ...any) {
	respHttp(w, respCode, nil, errS...)
}

func RespData(w http.ResponseWriter, data any, err error) {
	respHttp(w, http.StatusOK, data, "")
}
