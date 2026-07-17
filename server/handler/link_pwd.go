package handler

import (
	"bytes"
	"crypto/subtle"
	"html/template"
	"net/http"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
)

// 自助修改密码页面: 挂载在 VPN 服务(:443)的公开路由 /pwd 上, 无需管理员鉴权。
// 复用 dbdata.UserChangePassword 做核心校验, 复用 lockManager 做防爆限流。

const pwdCsrfCookie = "pwd-csrf"

// pwdPageData 用于渲染改密页面。Message 为结果提示, Success 决定提示样式。
type pwdPageData struct {
	CsrfToken string
	Message   string
	Success   bool
}

var pwdPageTpl = template.Must(template.New("pwd").Parse(pwdPageHTML))

// LinkPwd 处理 GET /pwd: 下发 CSRF token 并渲染表单。
func LinkPwd(w http.ResponseWriter, r *http.Request) {
	token := issuePwdCsrf(w)
	renderPwdPage(w, http.StatusOK, pwdPageData{CsrfToken: token})
}

// LinkPwdPost 处理 POST /pwd: 校验 CSRF、防爆限流, 调用核心逻辑修改密码。
func LinkPwdPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderPwdResult(w, false, "请求解析失败, 请重试")
		return
	}

	// CSRF 双提交校验: cookie 与表单字段必须一致。
	cookieToken, _ := GetCookie(r, pwdCsrfCookie)
	formToken := r.PostFormValue("csrf_token")
	if cookieToken == "" || subtle.ConstantTimeCompare([]byte(cookieToken), []byte(formToken)) != 1 {
		newToken := issuePwdCsrf(w)
		renderPwdPage(w, http.StatusOK, pwdPageData{CsrfToken: newToken, Message: "会话已过期, 请重新提交", Success: false})
		return
	}

	username := r.PostFormValue("username")
	oldPwd := r.PostFormValue("old_password")
	newPwd := r.PostFormValue("new_password")
	confirmPwd := r.PostFormValue("confirm_password")

	if newPwd != confirmPwd {
		renderPwdResult(w, false, "两次输入的新密码不一致")
		return
	}

	// 防爆限流: 与 VPN 登录共享 username+IP 锁定状态。
	if !lockManager.CheckLocked(username, r.RemoteAddr) {
		renderPwdResult(w, false, "尝试次数过多, 请稍后再试")
		return
	}

	err := dbdata.UserChangePassword(username, oldPwd, newPwd)
	if err != nil {
		// 旧密码错误计入防爆(用户名不存在也走同一路径, 避免枚举)。
		lockManager.UpdateLoginStatus(username, r.RemoteAddr, false)
		base.Warn("修改密码失败", username, r.RemoteAddr, err.Error())
		renderPwdResult(w, false, err.Error())
		return
	}

	// 成功后重置该 username+IP 的锁定计数。
	lockManager.UpdateLoginStatus(username, r.RemoteAddr, true)
	base.Info("用户自助修改密码成功", username, r.RemoteAddr)
	renderPwdResult(w, true, "密码修改成功, 请使用新密码登录 VPN")
}

// issuePwdCsrf 生成并下发一个新的 CSRF token(写入 cookie)。
func issuePwdCsrf(w http.ResponseWriter) string {
	token := utils.RandomRunes(32)
	SetCookie(w, pwdCsrfCookie, token, 0)
	return token
}

// renderPwdResult 渲染带结果提示的页面, 同时刷新 CSRF token 供再次提交。
func renderPwdResult(w http.ResponseWriter, success bool, msg string) {
	token := issuePwdCsrf(w)
	renderPwdPage(w, http.StatusOK, pwdPageData{CsrfToken: token, Message: msg, Success: success})
}

// renderPwdPage 输出 HTML 页面。
func renderPwdPage(w http.ResponseWriter, status int, data pwdPageData) {
	buf := new(bytes.Buffer)
	if err := pwdPageTpl.Execute(buf, data); err != nil {
		base.Error("渲染改密页面失败", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

const pwdPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>修改密码</title>
    <style>
        * { box-sizing: border-box; }
        body {
            margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center;
            background: #f0f2f5; color: #1f2329;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
        }
        .card { background: #fff; width: 100%; max-width: 380px; margin: 16px; padding: 32px 28px;
            border-radius: 10px; box-shadow: 0 4px 24px rgba(0,0,0,.08); }
        h1 { margin: 0 0 24px; font-size: 20px; text-align: center; }
        label { display: block; margin-bottom: 6px; font-size: 14px; color: #4e5969; }
        input { width: 100%; padding: 10px 12px; margin-bottom: 18px; border: 1px solid #d9d9d9;
            border-radius: 6px; font-size: 14px; }
        input:focus { outline: none; border-color: #165dff; }
        button { width: 100%; padding: 11px; background: #165dff; color: #fff; border: none;
            border-radius: 6px; font-size: 15px; cursor: pointer; }
        button:hover { background: #0e42d2; }
        .alert { padding: 10px 12px; border-radius: 6px; font-size: 14px; margin-bottom: 18px; }
        .alert-ok { background: #e8ffea; color: #009a29; border: 1px solid #aff0b5; }
        .alert-err { background: #ffece8; color: #cb2634; border: 1px solid #fdcdc5; }
        .hint { font-size: 12px; color: #86909c; margin-top: -10px; margin-bottom: 18px; }
    </style>
</head>
<body>
    <div class="card">
        <h1>修改密码</h1>
        {{if .Message}}
        <div class="alert {{if .Success}}alert-ok{{else}}alert-err{{end}}">{{.Message}}</div>
        {{end}}
        <form method="post" action="/pwd" autocomplete="off">
            <input type="hidden" name="csrf_token" value="{{.CsrfToken}}">
            <label for="username">用户名</label>
            <input type="text" id="username" name="username" required>
            <label for="old_password">旧密码</label>
            <input type="password" id="old_password" name="old_password" required>
            <div class="hint">仅需输入密码(PIN码), 无需附加动态码</div>
            <label for="new_password">新密码</label>
            <input type="password" id="new_password" name="new_password" required>
            <div class="hint">至少 6 位</div>
            <label for="confirm_password">确认新密码</label>
            <input type="password" id="confirm_password" name="confirm_password" required>
            <button type="submit">提交</button>
        </form>
    </div>
</body>
</html>`
