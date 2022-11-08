package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/skip2/go-qrcode"
)

func UserList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	prefix := r.FormValue("prefix")
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}

	var (
		pageSize = dbdata.PageSize
		count    int
		datas    []dbdata.User
		err      error
	)

	// 查询前缀匹配
	if len(prefix) > 0 {
		count = dbdata.CountPrefix("username", prefix, &dbdata.User{})
		err = dbdata.Prefix("username", prefix, &datas, pageSize, 1)
	} else {
		count = dbdata.CountAll(&dbdata.User{})
		err = dbdata.Find(&datas, pageSize, page)
	}

	if err != nil && !dbdata.CheckErrNotFound(err) {
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

func UserDetail(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	if id < 1 {
		RespError(w, RespParamErr, "用户名错误")
		return
	}

	var user dbdata.User
	err := dbdata.One("Id", id, &user)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	RespSucess(w, user)
}

func UserSet(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	data := &dbdata.User{}
	err = json.Unmarshal(body, data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	err = dbdata.SetUser(data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	// 发送邮件
	if data.SendEmail {
		err = userAccountMail(data)
		if err != nil {
			RespError(w, RespInternalErr, err)
			return
		}
	}
	//修改用户资料后执行过期用户检测
	sessdata.CloseUserLimittimeSession()
	RespSucess(w, nil)
}

func UserDel(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)

	if id < 1 {
		RespError(w, RespParamErr, "用户id错误")
		return
	}

	user := dbdata.User{Id: id}
	err := dbdata.Del(&user)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	RespSucess(w, nil)
}

func UserOtpQr(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	b64 := r.FormValue("b64")
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)
	var user dbdata.User
	err := dbdata.One("Id", id, &user)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	issuer := url.QueryEscape(base.Cfg.Issuer)
	qrstr := fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s", issuer, user.Email, issuer, user.OtpSecret)
	qr, _ := qrcode.New(qrstr, qrcode.High)

	if b64 == "1" {
		data, _ := qr.PNG(300)
		s := base64.StdEncoding.EncodeToString(data)
		_, err = fmt.Fprint(w, s)
		if err != nil {
			base.Error(err)
		}
		return
	}
	err = qr.Write(300, w)
	if err != nil {
		base.Error(err)
	}
}

// 在线用户
func UserOnline(w http.ResponseWriter, r *http.Request) {
	datas := sessdata.OnlineSess()

	data := map[string]interface{}{
		"count":     len(datas),
		"page_size": dbdata.PageSize,
		"datas":     datas,
	}

	RespSucess(w, data)
}

func UserOffline(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	token := r.FormValue("token")
	sessdata.CloseSess(token, dbdata.UserLogoutAdmin)
	RespSucess(w, nil)
}

func UserReline(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	token := r.FormValue("token")
	sessdata.CloseCSess(token)
	RespSucess(w, nil)
}

type userAccountMailData struct {
	Issuer   string
	LinkAddr string
	Group    string
	Username string
	PinCode  string
	OtpImg   string
}

func userAccountMail(user *dbdata.User) error {
	// 平台通知
	htmlBody := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <title>Hello AnyLink!</title>
</head>
<body>
%s
</body>
</html>
`
	dataOther := &dbdata.SettingOther{}
	err := dbdata.SettingGet(dataOther)
	if err != nil {
		base.Error(err)
		return err
	}
	htmlBody = fmt.Sprintf(htmlBody, dataOther.AccountMail)
	// fmt.Println(htmlBody)

	// token有效期3天
	expiresAt := time.Now().Unix() + 3600*24*3
	jwtData := map[string]interface{}{"id": user.Id}
	tokenString, err := SetJwtData(jwtData, expiresAt)
	if err != nil {
		return err
	}

	setting := &dbdata.SettingOther{}
	err = dbdata.SettingGet(setting)
	if err != nil {
		base.Error(err)
		return err
	}

	data := userAccountMailData{
		LinkAddr: setting.LinkAddr,
		Group:    strings.Join(user.Groups, ","),
		Username: user.Username,
		PinCode:  user.PinCode,
		OtpImg:   fmt.Sprintf("https://%s/otp_qr?id=%d&jwt=%s", setting.LinkAddr, user.Id, tokenString),
	}
	w := bytes.NewBufferString("")
	t, _ := template.New("auth_complete").Parse(htmlBody)
	err = t.Execute(w, data)
	if err != nil {
		return err
	}
	// fmt.Println(w.String())
	return SendMail(base.Cfg.Issuer+"平台通知", user.Email, w.String())
}
