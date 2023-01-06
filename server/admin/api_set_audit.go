package admin

import (
	"net/http"
	"strconv"

	"github.com/bjdgyc/anylink/dbdata"
	"github.com/gocarina/gocsv"
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

func SetAuditExport(w http.ResponseWriter, r *http.Request) {
	var datas []dbdata.AccessAudit
	maxNum := 1000000
	session := dbdata.GetAuditSession(r.FormValue("search"))
	count, err := dbdata.FindAndCount(session, &datas, maxNum, 0)
	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}
	if count == 0 {
		RespError(w, RespParamErr, "你导出的总数量为0条，请调整搜索条件，再导出")
		return
	}
	if count > int64(maxNum) {
		RespError(w, RespParamErr, "你导出的数据量超过100万条，请调整搜索条件，再导出")
		return
	}
	gocsv.Marshal(datas, w)

}

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
