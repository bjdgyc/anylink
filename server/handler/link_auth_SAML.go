package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
)

func getServerAddr(r *http.Request) string {
	return "https://" + r.Host
}
func SAMLSPLogin(w http.ResponseWriter, r *http.Request) {
	tgname := r.URL.Query().Get("tgname")
	if tgname == "" {
		base.Error("缺少组名参数")
		return
	}

	// 生成临时会话 ID
	sessionID, err := GenerateSessionID()
	if err != nil {
		base.Error("临时会话创建失败", err)
		return
	}

	// 创建临时会话（用于传递组名）
	tempSession := &AuthSession{
		ClientRequest: &ClientRequest{
			GroupSelect: tgname,
		},
	}

	// 保存临时会话
	SessStore.SaveAuthSession(sessionID, tempSession)
	// 获取企微配置
	wxworkConfig, err := dbdata.GetAuthWework(tgname)
	if err != nil {
		base.Error("获取企微配置失败", err)
		return
	}
	corpId := wxworkConfig.CorpId
	agentId := wxworkConfig.AgentId

	// 企微认证回调地址
	redirectUri := fmt.Sprintf("%s/WXAuth/callback", getServerAddr(r))

	weworkUrl := fmt.Sprintf("https://login.work.weixin.qq.com/wwlogin/sso/login?login_type=CorpApp&appid=%s&agentid=%s&redirect_uri=%s&state=%s",
		corpId, agentId, url.QueryEscape(redirectUri), url.QueryEscape(sessionID), // 使用state传递临时sessionID
	)
	// 重定向到企业微信扫码页面
	http.Redirect(w, r, weworkUrl, http.StatusFound)
}

func WXAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	tempSessionID := r.URL.Query().Get("state") // 通过state参数获取临时会话ID

	if code == "" || tempSessionID == "" {
		base.Error("企微认证回调缺少参数")
		return
	}

	// 获取临时会话
	tempSession, err := SessStore.GetAuthSession(tempSessionID)
	if err != nil {
		base.Error("无效或过期的会话")
		return
	}
	groupname := tempSession.ClientRequest.GroupSelect
	// 获取企微配置
	wxworkConfig, err := dbdata.GetAuthWework(groupname)
	if err != nil {
		base.Error("获取企微配置失败", err)
		return
	}
	// 调用企业微信 API 获取用户信息
	user := dbdata.AuthWXwork{}
	userID, err := user.GetWeworkUser(wxworkConfig.CorpId, wxworkConfig.Secret, code)
	if err != nil {
		base.Error("用户信息获取失败", err)
		return
	}
	username := userID

	// 创建SAML会话 用于传递组名和用户名
	samlSession := &AuthSession{
		ClientRequest: &ClientRequest{
			GroupSelect: tempSession.ClientRequest.GroupSelect,
			Auth: auth{
				Username: username,
			},
		},
	}
	// 保存saml会话
	SessStore.SaveAuthSession(code, samlSession)
	// 删除临时会话
	SessStore.DeleteAuthSession(tempSessionID)

	// 设置 Cookie
	SetCookie(w, "acSamlv2Token", code, 0)

	// 重定向到 sso-v2-login-final URL（需严格符合cisco anyconnect路由格式，不允许带任何参数！）
	http.Redirect(w, r, "/+CSCOE+/saml_ac_login.html", http.StatusFound)
}

// SAML回调端点 - 对应 sso-v2-login-final
func SAMLACLogin(w http.ResponseWriter, r *http.Request) {
	// 验证 Cookie 是否存在
	if token, err := GetCookie(r, "acSamlv2Token"); err != nil || token == "" {
		base.Error("认证信息丢失,获取Cookie失败")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	html := `
			<!DOCTYPE html>
			<html>
			<head><title>认证成功</title></head>
			<body>
				<p>身份验证成功，正在建立 VPN 连接...</p>
				<p>请稍候，窗口即将关闭。</p>
			</body>
			</html>
		`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func SAMLTest(w http.ResponseWriter, r *http.Request) {
	var test = "8lmahgBycwtHwwIN"
	w.Write([]byte(test))
}

var auth_request_saml = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request" aggregate-auth-version="2">
    <opaque is-for="sg">
        <tunnel-group>{{.Group}}</tunnel-group>
        <group-alias>{{.Group}}</group-alias>
        <aggauth-handle>168179266</aggauth-handle>
        <config-hash>1595829378234</config-hash>
        <auth-method>single-sign-on-v2</auth-method>
    </opaque>
    <auth id="main">
        <title>SAML SSO Login</title>
        <message>请完成SAML单点登录认证</message>
        <banner></banner>
        {{if .Error}}
        <error id="88" param1="{{.Error}}" param2="">SAML认证失败: %s</error>
        {{end}}
        <sso-v2-login>{{.ServerAddr}}/+CSCOE+/saml/sp/login?tgname={{.Group}}&#x26;acsamlcap=v2</sso-v2-login>
        <sso-v2-login-final>{{.ServerAddr}}/+CSCOE+/saml_ac_login.html</sso-v2-login-final>
        <sso-v2-token-cookie-name>acSamlv2Token</sso-v2-token-cookie-name>
        <sso-v2-error-cookie-name>acSamlv2Error</sso-v2-error-cookie-name>
        <form>
            <input type="sso" name="sso-token"></input>
        </form>
    </auth>
</config-auth>`

// <sso-v2-browser-mode>external</sso-v2-browser-mode>
