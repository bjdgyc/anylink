package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
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

	wxWorkUrl := fmt.Sprintf("https://login.work.weixin.qq.com/wwlogin/sso/login?login_type=CorpApp&appid=%s&agentid=%s&redirect_uri=%s&state=%s",
		corpId, agentId, url.QueryEscape(redirectUri), url.QueryEscape(utils.RandomRunes(32)+tgname), // 使用state传递组名,添加随机字符防止CSRF 攻击
	)
	// 重定向到企业微信扫码页面
	http.Redirect(w, r, wxWorkUrl, http.StatusFound)
}

func WXAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state") // 通过state参数获取组名

	if code == "" || state == "" {
		base.Error("企微认证回调缺少参数")
		return
	}
	groupname := state[32:]
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
			GroupSelect: groupname,
			Auth: auth{
				Username: username,
			},
		},
	}
	// 保存saml会话
	SessStore.SaveAuthSession(code, samlSession)

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
<head>
	<title>认证成功</title>
	<style>
		body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
		.success { color: green; font-size: 24px; }
	</style>
</head>
<body>
	<div class="success">
		<h1>认证成功</h1>
		<p>您已成功通过认证，请关闭此浏览器窗口并返回VPN客户端。</p>
		<p>VPN客户端将自动完成连接过程。</p>
	</div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func SAMLTest(w http.ResponseWriter, r *http.Request) {
	var test = "8lmahgBycwtHwwIN"
	w.Write([]byte(test))
}

// var auth_reply_saml = `<?xml version='1.0' encoding='UTF-8'?>
// <config-auth client="vpn" type="auth-reply" aggregate-auth-version="2">
//   <version who="vpn">4.7.00136</version>
//   <device-id>linux-64</device-id>
//   <session-token/>
//   <session-id/>
//   <opaque is-for="sg">
//     <tunnel-group>DefaultWEBVPNGroup</tunnel-group>
//     <auth-method>single-sign-on-v2</auth-method>
//     <config-hash>1646156124329</config-hash>
//   </opaque>
//   <auth>
//     <sso-token>71D2D28F0744FDEE74C74F6</sso-token>
//   </auth>
// </config-auth>
// `

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
        <form>
            <input type="sso" name="sso-token"></input>
        </form>
    </auth>
</config-auth>`

// <sso-v2-error-cookie-name>acSamlv2Error</sso-v2-error-cookie-name>
// <sso-v2-browser-mode>external</sso-v2-browser-mode>
// <sso-token>{{.SsoToken}}</sso-token>
