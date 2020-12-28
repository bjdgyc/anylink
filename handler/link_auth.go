package handler

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkAuth(w http.ResponseWriter, r *http.Request) {
	// 判断anyconnect客户端
	userAgent := strings.ToLower(r.UserAgent())
	x_Aggregate_Auth := r.Header.Get("X-Aggregate-Auth")
	x_Transcend_Version := r.Header.Get("X-Transcend-Version")
	if !(strings.Contains(userAgent, "anyconnect") &&
		x_Aggregate_Auth == "1" && x_Transcend_Version == "1") {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "error request")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	cr := ClientRequest{}
	err = xml.Unmarshal(body, &cr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Printf("%+v \n", cr)

	setCommonHeader(w)
	if cr.Type == "logout" {
		// 退出删除session信息
		if cr.SessionToken != "" {
			sessdata.DelSessByStoken(cr.SessionToken)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if cr.Type == "init" {
		w.WriteHeader(http.StatusOK)
		data := RequestData{Group: cr.GroupSelect, Groups: dbdata.GetGroupNames()}
		tplRequest(tpl_request, w, data)
		return
	}

	// 登陆参数判断
	if cr.Type != "auth-reply" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO 用户密码校验
	err = dbdata.CheckUser(cr.Auth.Username, cr.Auth.Password, cr.GroupSelect)
	if err != nil {
		base.Info(err)
		w.WriteHeader(http.StatusOK)
		data := RequestData{Group: cr.GroupSelect, Groups: dbdata.GetGroupNames(), Error: "用户名或密码错误"}
		tplRequest(tpl_request, w, data)
		return
	}
	// if !ok {
	//	w.WriteHeader(http.StatusOK)
	//	data := RequestData{Group: cr.GroupSelect, Groups: base.Cfg.UserGroups, Error: "请先激活用户"}
	//	tplRequest(tpl_request, w, data)
	//	return
	// }

	// 创建新的session信息
	sess := sessdata.NewSession("")
	sess.Username = cr.Auth.Username
	sess.Group = cr.GroupSelect
	sess.MacAddr = strings.ToLower(cr.MacAddressList.MacAddress)
	sess.UniqueIdGlobal = cr.DeviceId.UniqueIdGlobal
	other := &dbdata.SettingOther{}
	dbdata.SettingGet(other)
	rd := RequestData{SessionId: sess.Sid, SessionToken: sess.Sid + "@" + sess.Token,
		Banner: other.Banner}
	w.WriteHeader(http.StatusOK)
	tplRequest(tpl_complete, w, rd)
}

const (
	tpl_request = iota
	tpl_complete
)

func tplRequest(typ int, w io.Writer, data RequestData) {
	if typ == tpl_request {
		t, _ := template.New("auth_request").Parse(auth_request)
		t.Execute(w, data)
		return
	}

	if strings.Contains(data.Banner, "\n") {
		// 替换xml文件的换行符
		data.Banner = strings.ReplaceAll(data.Banner, "\n", "&#x0A;")
	}
	t, _ := template.New("auth_complete").Parse(auth_complete)
	t.Execute(w, data)
}

// 设置输出信息
type RequestData struct {
	Groups []string
	Group  string
	Error  string

	// complete
	SessionId    string
	SessionToken string
	Banner       string
}

var auth_request = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request" aggregate-auth-version="2">
    <opaque is-for="sg">
        <tunnel-group>{{.Group}}</tunnel-group>
        <group-alias>{{.Group}}</group-alias>
        <aggauth-handle>168179266</aggauth-handle>
        <config-hash>1595829378234</config-hash>
        <auth-method>multiple-cert</auth-method>
        <auth-method>single-sign-on-v2</auth-method>
    </opaque>
    <auth id="main">
        <title>Login</title>
        <message>请输入你的用户名和密码</message>
        <banner></banner>
        {{if .Error}}
        <error id="88" param1="{{.Error}}" param2="">登陆失败:  %s</error>
        {{end}}
        <form>
            <input type="text" name="username" label="Username:"></input>
            <input type="password" name="password" label="Password:"></input>
            <select name="group_list" label="GROUP:">
                {{range $v := .Groups}}
                <option {{if eq $v $.Group}} selected="true"{{end}}>{{$v}}</option>
                {{end}}
            </select>
        </form>
    </auth>
</config-auth>
`

var auth_complete = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="complete" aggregate-auth-version="2">
    <session-id>{{.SessionId}}</session-id>
    <session-token>{{.SessionToken}}</session-token>
    <auth id="success">
        <banner>{{.Banner}}</banner>
        <message id="0" param1="" param2=""></message>
    </auth>
    <capabilities>
        <crypto-supported>ssl-dhe</crypto-supported>
    </capabilities>
    <config client="vpn" type="private">
        <vpn-base-config>
            <server-cert-hash>240B97A685B2BFA66AD699B90AAC49EA66495D69</server-cert-hash>
        </vpn-base-config>
        <opaque is-for="vpn-client"></opaque>
    </config>
</config-auth>
`
