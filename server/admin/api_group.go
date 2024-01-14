package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/bjdgyc/anylink/dbdata"
)

func GroupList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}

	var pageSize = dbdata.PageSize

	count := dbdata.CountAll(&dbdata.Group{})

	var datas []dbdata.Group
	err := dbdata.Find(&datas, pageSize, page)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	data := map[string]any{
		"count":     count,
		"page_size": pageSize,
		"datas":     datas,
	}

	RespSucess(w, data)
}

func GroupNames(w http.ResponseWriter, r *http.Request) {
	var names = dbdata.GetGroupNames()
	data := map[string]any{
		"count":     len(names),
		"page_size": 0,
		"datas":     names,
	}
	RespSucess(w, data)
}

func GroupNamesIds(w http.ResponseWriter, r *http.Request) {
	var names = dbdata.GetGroupNamesIds()
	data := map[string]any{
		"count":     len(names),
		"page_size": 0,
		"datas":     names,
	}
	RespSucess(w, data)
}

func GroupDetail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	if id < 1 {
		RespError(w, RespParamErr, "Id错误")
		return
	}

	var data dbdata.Group
	err := dbdata.One("Id", id, &data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if len(data.Auth) == 0 {
		data.Auth["type"] = "local"
	}
	RespSucess(w, data)
}

func GroupSet(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	v := &dbdata.Group{}
	err = json.Unmarshal(body, v)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	err = dbdata.SetGroup(v)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	RespSucess(w, nil)
}

func GroupDel(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	if id < 1 {
		RespError(w, RespParamErr, "Id错误")
		return
	}

	data := dbdata.Group{Id: id}
	err := dbdata.Del(&data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	RespSucess(w, nil)
}

func GroupAuthLogin(w http.ResponseWriter, r *http.Request) {
	type AuthLoginData struct {
		Name string         `json:"name"`
		Pwd  string         `json:"pwd"`
		Auth map[string]any `json:"auth"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	v := &AuthLoginData{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	err = dbdata.GroupAuthLogin(v.Name, v.Pwd, v.Auth)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	RespSucess(w, "ok")
}
