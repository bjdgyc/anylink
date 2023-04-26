package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/bjdgyc/anylink/dbdata"
)

func UserIpMapList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}

	var pageSize = dbdata.PageSize

	count := dbdata.CountAll(&dbdata.IpMap{})

	var datas []dbdata.IpMap
	err := dbdata.Find(&datas, pageSize, page)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	data := map[string]interface{}{
		"count":     count,
		"page_size": pageSize,
		"datas":     datas,
	}

	RespSucess(w, data)
}

func UserIpMapDetail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	if id < 1 {
		RespError(w, RespParamErr, "用户名错误")
		return
	}

	var data dbdata.IpMap
	err := dbdata.One("Id", id, &data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	RespSucess(w, data)
}

func UserIpMapSet(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	v := &dbdata.IpMap{}
	err = json.Unmarshal(body, v)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	// fmt.Println(v, len(v.Ip), len(v.MacAddr))

	err = dbdata.SetIpMap(v)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	// sessdata.IpAllSet(v)

	RespSucess(w, nil)
}

func UserIpMapDel(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)

	if id < 1 {
		RespError(w, RespParamErr, "IP映射id错误")
		return
	}

	var data dbdata.IpMap
	err := dbdata.One("Id", id, &data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	err = dbdata.Del(&data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	// sessdata.IpAllDel(&data)

	RespSucess(w, nil)
}
