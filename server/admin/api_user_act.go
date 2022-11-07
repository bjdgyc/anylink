package admin

import (
	"net/http"
	"strconv"

	"github.com/bjdgyc/anylink/dbdata"
)

func UserActLogList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}
	var datas []dbdata.UserActLog
	session := dbdata.UserActLogIns.GetSession(r.Form)
	count, err := dbdata.FindAndCount(session, &datas, dbdata.PageSize, page)
	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}
	data := map[string]interface{}{
		"count":     count,
		"page_size": dbdata.PageSize,
		"datas":     datas,
		"statusOps": dbdata.UserActLogIns.GetStatusOpsWithTag(),
		"osOps":     dbdata.UserActLogIns.OsOps,
		"clientOps": dbdata.UserActLogIns.ClientOps,
	}

	RespSucess(w, data)
}
