package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/skip2/go-qrcode"
	mail "github.com/xhit/go-simple-mail/v2"
)

func UserList(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	prefix := r.FormValue("prefix")
	prefix = strings.TrimSpace(prefix)
	pageS := r.FormValue("page")
	page, _ := strconv.Atoi(pageS)
	if page < 1 {
		page = 1
	}

	// 添加从请求参数读取 page_size
	pageSizeS := r.FormValue("page_size")
	pageSize, _ := strconv.Atoi(pageSizeS)
	if pageSize <= 0 {
		pageSize = dbdata.PageSize // 使用默认值
	}

	var (
		count int
		datas []dbdata.User
		err   error
	)

	// 查询前缀匹配
	if len(prefix) > 0 {
		fuzzy := "%" + prefix + "%"
		where := "username LIKE ? OR nickname LIKE ? OR email LIKE ? OR type LIKE ?"

		count = dbdata.FindWhereCount(&dbdata.User{}, where, fuzzy, fuzzy, fuzzy, fuzzy)
		err = dbdata.FindWhere(&datas, pageSize, page, where, fuzzy, fuzzy, fuzzy, fuzzy)
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

	if len(data.PinCode) < 6 {
		data.PinCode = utils.RandomRunes(8)
		base.Info("用户", data.Username, "随机密码为:", data.PinCode)
	}
	plainpwd := data.PinCode
	err = dbdata.SetUser(data)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	data.PinCode = plainpwd

	// 发送邮件
	if data.SendEmail {
		err = userAccountMail(data)
		if err != nil {
			RespError(w, RespInternalErr, err)
			return
		}
	}
	// 修改用户资料后执行过期用户检测
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
	b64S := r.FormValue("b64")
	idS := r.FormValue("id")
	id, _ := strconv.Atoi(idS)

	var b64 bool
	if b64S == "1" {
		b64 = true
	}
	data, err := userOtpQr(id, b64)
	if err != nil {
		base.Error(err)
	}
	io.WriteString(w, data)
}

func userOtpQr(uid int, b64 bool) (string, error) {
	var user dbdata.User
	err := dbdata.One("Id", uid, &user)
	if err != nil {
		return "", err
	}

	issuer := url.QueryEscape(base.Cfg.Issuer)
	qrstr := fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s", issuer, user.Email, issuer, user.OtpSecret)
	qr, _ := qrcode.New(qrstr, qrcode.High)

	if b64 {
		data, err := qr.PNG(300)
		if err != nil {
			return "", err
		}
		s := base64.StdEncoding.EncodeToString(data)
		return s, nil
	}

	buf := bytes.NewBuffer(nil)
	err = qr.Write(300, buf)
	return buf.String(), err
}

// 在线用户
func UserOnline(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	search_cate := r.FormValue("search_cate")
	search_text := r.FormValue("search_text")
	show_sleeper := r.FormValue("show_sleeper")
	showSleeper, _ := strconv.ParseBool(show_sleeper)
	// one_offline := r.FormValue("one_offline")

	// datas := sessdata.OnlineSess()
	datas := sessdata.GetOnlineSess(search_cate, search_text, showSleeper)

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

// 批量发送邮件
func UserBatchSendEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserIds []int `json:"user_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	if len(req.UserIds) == 0 {
		RespError(w, RespInternalErr, errors.New("用户ID列表不能为空"))
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	failCount := 0

	// 限制最大并发数
	maxConcurrency := min(len(req.UserIds), 10)
	concurrencyLimiter := make(chan struct{}, maxConcurrency)

	for _, userId := range req.UserIds {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()
			concurrencyLimiter <- struct{}{}
			defer func() { <-concurrencyLimiter }()

			user := &dbdata.User{}
			err := dbdata.One("Id", uid, user)
			if err != nil {
				base.Error("批量发送邮件失败，获取用户信息错误:", uid, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			err = userAccountMail(user)
			mu.Lock()
			if err != nil {
				base.Error("批量发送邮件失败:", user.Username, err)
				failCount++
			} else {
				successCount++
			}
			mu.Unlock()
		}(userId)
	}

	wg.Wait()

	msg := fmt.Sprintf("批量发送邮件完成，成功：%d，失败：%d", successCount, failCount)

	if successCount > 0 {
		RespSucess(w, msg)
	} else {
		RespError(w, RespInternalErr, errors.New(msg))
	}
}

// 批量删除用户
func UserBatchDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserIds []int `json:"user_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	if len(req.UserIds) == 0 {
		RespError(w, RespInternalErr, errors.New("用户ID列表不能为空"))
		return
	}

	successCount := 0
	failCount := 0

	for _, userId := range req.UserIds {
		user := &dbdata.User{}
		err := dbdata.One("Id", userId, user)
		if err != nil {
			failCount++
			continue
		}

		err = dbdata.Del(user)
		if err != nil {
			base.Error("批量删除用户失败:", user.Username, err)
			failCount++
		} else {
			successCount++
		}
	}

	msg := fmt.Sprintf("批量删除完成，成功：%d，失败：%d", successCount, failCount)

	if successCount > 0 {
		RespSucess(w, msg)
	} else {
		RespError(w, RespInternalErr, errors.New(msg))
	}
}

type userAccountMailData struct {
	Issuer       string
	LinkAddr     string
	Group        string
	Username     string
	Nickname     string
	PinCode      string
	LimitTime    string
	OtpImg       string
	OtpImgBase64 string
	DisableOtp   bool
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

	otpData, _ := userOtpQr(user.Id, true)

	data := userAccountMailData{
		Issuer:       base.Cfg.Issuer,
		LinkAddr:     setting.LinkAddr,
		Group:        strings.Join(user.Groups, ","),
		Username:     user.Username,
		Nickname:     user.Nickname,
		PinCode:      user.PinCode,
		OtpImg:       fmt.Sprintf("https://%s/otp_qr?id=%d&jwt=%s", setting.LinkAddr, user.Id, tokenString),
		OtpImgBase64: "data:image/png;base64," + otpData,
		DisableOtp:   user.DisableOtp,
	}

	if user.Type == "ldap" {
		data.PinCode = "同ldap密码"
	}

	if user.LimitTime == nil {
		data.LimitTime = "无限制"
	} else {
		data.LimitTime = user.LimitTime.Local().Format("2006-01-02")
	}

	w := bytes.NewBufferString("")
	t, _ := template.New("auth_complete").Parse(htmlBody)
	err = t.Execute(w, data)
	if err != nil {
		return err
	}
	// fmt.Println(w.String())

	var otpAttach *mail.File
	if !user.DisableOtp {
		imgData, _ := userOtpQr(user.Id, false)
		otpAttach = &mail.File{
			MimeType: "image/png",
			Name:     "userOtpQr.png",
			Data:     []byte(imgData),
			Inline:   true,
		}
	}

	// 自动生成客户端证书并作为附件发送(证书已存在则复用)
	var p12Attach *mail.File
	if len(user.Groups) > 0 {
		groupname := user.Groups[0]
		if _, err := dbdata.GetClientCert(user.Username, groupname); err != nil {
			if _, err := dbdata.GenerateClientCert(user.Username, groupname, false, 0); err != nil {
				base.Warn("自动生成客户端证书失败:", user.Username, err)
			}
		}
		p12Pass := utils.RandomRunes(8)
		if p12Data, err := dbdata.GenerateClientP12FromDB(user.Username, groupname, p12Pass); err == nil {
			p12Name := fmt.Sprintf("vpn-%s.p12", user.Username)
			p12Attach = &mail.File{
				MimeType: "application/x-pkcs12",
				Name:     p12Name,
				Data:     p12Data,
			}
			// 邮件正文追加证书导入说明
			certSection := fmt.Sprintf(`
<hr/>
<h3>证书登录信息</h3>
<p>本邮件附件包含 VPN 客户端证书 <b>%s</b>，安装后无需账号密码即可登录。</p>
<ul>
<li>P12 导入密码：<b>%s</b></li>
<li>所属用户组：%s</li>
<li>证书有效期：1 年，到期请联系管理员</li>
</ul>
<p>导入方式：Cisco Secure Client 打开 首选项 → 证书 → 导入，选择附件并输入上面的密码；iOS/Android 设备直接点开附件安装，然后在 App 的证书设置中选中。</p>
<p><b>请妥善保管此邮件，P12 文件等同于登录凭证。</b></p>
`, p12Name, p12Pass, groupname)
			w.WriteString(certSection)
		} else {
			base.Warn("生成客户端P12失败:", user.Username, err)
		}
	}

	return SendMail(base.Cfg.Issuer, user.Email, w.String(), otpAttach, p12Attach)
}
