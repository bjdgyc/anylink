package handler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ClientRequest struct {
	XMLName              xml.Name       `xml:"config-auth"`
	Client               string         `xml:"client,attr"`                 // 一般都是 vpn
	Type                 string         `xml:"type,attr"`                   // 请求类型 init logout auth-reply
	AggregateAuthVersion string         `xml:"aggregate-auth-version,attr"` // 一般都是 2
	Version              string         `xml:"version"`                     // 客户端版本号
	GroupAccess          string         `xml:"group-access"`                // 请求的地址
	GroupSelect          string         `xml:"group-select"`                // 选择的组名
	SessionId            string         `xml:"session-id"`
	SessionToken         string         `xml:"session-token"`
	Auth                 auth           `xml:"auth"`
	DeviceId             deviceId       `xml:"device-id"`
	MacAddressList       macAddressList `xml:"mac-address-list"`
}

type auth struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
}

type deviceId struct {
	ComputerName    string `xml:"computer-name,attr"`
	DeviceType      string `xml:"device-type,attr"`
	PlatformVersion string `xml:"platform-version,attr"`
	UniqueId        string `xml:"unique-id,attr"`
	UniqueIdGlobal  string `xml:"unique-id-global,attr"`
}

type macAddressList struct {
	MacAddress string `xml:"mac-address"`
}

// 判断anyconnect客户端
func checkVpnClient(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// TODO 调试信息输出
		// hd, _ := httputil.DumpRequest(r, true)
		// fmt.Println("DumpRequest: ", string(hd))
		fmt.Println(r.RemoteAddr)

		user_Agent := strings.ToLower(r.UserAgent())
		x_Aggregate_Auth := r.Header.Get("X-Aggregate-Auth")
		x_Transcend_Version := r.Header.Get("X-Transcend-Version")
		if strings.Contains(user_Agent, "anyconnect") &&
			x_Aggregate_Auth == "1" && x_Transcend_Version == "1" {
			h(w, r, ps)
		} else {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "error request")
		}
	}
}

func setCommonHeader(w http.ResponseWriter) {
	// Content-Length Date 默认已经存在
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Aggregate-Auth", "1")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}
