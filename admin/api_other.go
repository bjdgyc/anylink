package admin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/bjdgyc/anylink/dbdata"
)

func setOtherGet(data interface{}, w http.ResponseWriter) {
	err := dbdata.SettingGet(data)
	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}
	RespSucess(w, data)
}

func setOtherEdit(data interface{}, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	// fmt.Println(data)

	err = dbdata.SettingSet(data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	RespSucess(w, data)
}

func SetOtherSmtp(w http.ResponseWriter, r *http.Request) {
	data := &dbdata.SettingSmtp{}
	setOtherGet(data, w)
}

func SetOtherSmtpEdit(w http.ResponseWriter, r *http.Request) {
	data := &dbdata.SettingSmtp{}
	setOtherEdit(data, w, r)
}

func SetOther(w http.ResponseWriter, r *http.Request) {
	data := &dbdata.SettingOther{}
	setOtherGet(data, w)
}

func SetOtherEdit(w http.ResponseWriter, r *http.Request) {
	data := &dbdata.SettingOther{}
	setOtherEdit(data, w, r)
}
