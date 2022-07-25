package admin

import (
	"net/http"
	"strconv"

	"github.com/bjdgyc/anylink/dbdata"
)

func SetAuditList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}
	var datas []dbdata.AccessAudit
	session := dbdata.GetAuditSession(r.FormValue("search"))
	count, err := dbdata.FindAndCount(session, &datas, dbdata.PageSize, page)
	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}
	data := map[string]interface{}{
		"count":     count,
		"page_size": dbdata.PageSize,
		"datas":     datas,
	}

	RespSucess(w, data)
}
