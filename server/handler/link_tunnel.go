package handler

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
)

var (
	hn string
)

func init() {
	// 获取主机名称
	hn, _ = os.Hostname()
}

func Http_SetHeader(w http.ResponseWriter, key string, value string) {
   w.Header()[key] = []string{value}
}

func Http_AddHeader(w http.ResponseWriter, key string, value string) {
   w.Header()[key] = append(w.Header()[key], value)
}

func LinkTunnel(w http.ResponseWriter, r *http.Request) {
	// TODO 调试信息输出
	// hd, _ := httputil.DumpRequest(r, true)
	// fmt.Println("DumpRequest: ", string(hd))
	// fmt.Println("LinkTunnel", r.RemoteAddr)

	// 判断session-token的值
	cookie, err := r.Cookie("webvpn")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sess := sessdata.SToken2Sess(cookie.Value)
	if sess == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 开启link
	cSess := sess.NewConn()
	if cSess == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 客户端信息
	cstpMtu := r.Header.Get("X-CSTP-MTU")
	cstpBaseMtu := r.Header.Get("X-CSTP-Base-MTU")
	masterSecret := r.Header.Get("X-DTLS-Master-Secret")
	localIp := r.Header.Get("X-Cstp-Local-Address-Ip4")
	mobile := r.Header.Get("X-Cstp-License")

	cSess.SetMtu(cstpMtu)
	cSess.MasterSecret = masterSecret
	cSess.RemoteAddr = r.RemoteAddr
	cSess.LocalIp = net.ParseIP(localIp)
	cstpKeepalive := base.Cfg.CstpKeepalive
	cstpDpd := base.Cfg.CstpDpd
	cSess.Client = "pc"
	if mobile == "mobile" {
		// 手机客户端
		cstpKeepalive = base.Cfg.MobileKeepalive
		cstpDpd = base.Cfg.MobileDpd
		cSess.Client = "mobile"
	}
	cSess.CstpDpd = cstpDpd

	dtlsPort := "4433"
	if strings.Contains(base.Cfg.ServerDTLSAddr, ":") {
		ss := strings.Split(base.Cfg.ServerDTLSAddr, ":")
		dtlsPort = ss[1]
	}

	base.Debug(cSess.IpAddr, cSess.MacHw, sess.Username, mobile)

	// 返回客户端数据
	Http_SetHeader(w, "Server", fmt.Sprintf("%s %s", base.APP_NAME, base.APP_VER))
	Http_SetHeader(w, "X-CSTP-Version", "1")
	Http_SetHeader(w, "X-CSTP-Server-Name", fmt.Sprintf("%s %s", base.APP_NAME, base.APP_VER))
	Http_SetHeader(w, "X-CSTP-Protocol", "Copyright (c) 2004 Cisco Systems, Inc.")
	Http_SetHeader(w, "X-CSTP-Address", cSess.IpAddr.String())             // 分配的ip地址
	Http_SetHeader(w, "X-CSTP-Netmask", sessdata.IpPool.Ipv4Mask.String()) // 子网掩码
	Http_SetHeader(w, "X-CSTP-Hostname", hn)                               // 机器名称
	//Http_SetHeader(w, "X-CSTP-Default-Domain", cSess.LocalIp)          
	Http_SetHeader(w, "X-CSTP-Base-MTU", cstpBaseMtu)

	// 允许本地LAN访问vpn网络，必须放在路由的第一个
	if cSess.Group.AllowLan {
		Http_SetHeader(w, "X-CSTP-Split-Exclude", "0.0.0.0/255.255.255.255")
	}
	// dns地址
	for _, v := range cSess.Group.ClientDns {
		Http_AddHeader(w, "X-CSTP-DNS", v.Val)
	}
	// 允许的路由
	for _, v := range cSess.Group.RouteInclude {
		if v.Val == "all" {
			continue
		}
		Http_AddHeader(w, "X-CSTP-Split-Include", v.IpMask)
	}
	// 不允许的路由
	for _, v := range cSess.Group.RouteExclude {
		Http_AddHeader(w, "X-CSTP-Split-Exclude", v.IpMask)
	}

	Http_SetHeader(w, "X-CSTP-Lease-Duration", fmt.Sprintf("%d", base.Cfg.IpLease)) // ip地址租期
	Http_SetHeader(w, "X-CSTP-Session-Timeout", "none")
	Http_SetHeader(w, "X-CSTP-Session-Timeout-Alert-Interval", "60")
	Http_SetHeader(w, "X-CSTP-Session-Timeout-Remaining", "none")
	Http_SetHeader(w, "X-CSTP-Idle-Timeout", "18000")
	Http_SetHeader(w, "X-CSTP-Disconnected-Timeout", "18000")
	Http_SetHeader(w, "X-CSTP-Keep", "true")
	Http_SetHeader(w, "X-CSTP-Tunnel-All-DNS", "false")

	Http_SetHeader(w, "X-CSTP-Rekey-Time", "172800")
	Http_SetHeader(w, "X-CSTP-Rekey-Method", "new-tunnel")

	Http_SetHeader(w, "X-CSTP-DPD", fmt.Sprintf("%d", cstpDpd))
	Http_SetHeader(w, "X-CSTP-Keepalive", fmt.Sprintf("%d", cstpKeepalive))
	// Http_SetHeader(w, "X-CSTP-Banner", banner.Banner)
	Http_SetHeader(w, "X-CSTP-MSIE-Proxy-Lockdown", "true")
	Http_SetHeader(w, "X-CSTP-Smartcard-Removal-Disconnect", "true")

	Http_SetHeader(w, "X-CSTP-MTU", fmt.Sprintf("%d", cSess.Mtu)) // 1399
	Http_SetHeader(w, "X-DTLS-MTU", fmt.Sprintf("%d", cSess.Mtu))

	Http_SetHeader(w, "X-DTLS-Session-ID", sess.DtlsSid)
	Http_SetHeader(w, "X-DTLS-Port", dtlsPort)
	Http_SetHeader(w, "X-DTLS-DPD", fmt.Sprintf("%d", cstpDpd))
	Http_SetHeader(w, "X-DTLS-Keepalive", fmt.Sprintf("%d", cstpKeepalive))
	Http_SetHeader(w, "X-DTLS-Rekey-Time", "5400")
	Http_SetHeader(w, "X-DTLS12-CipherSuite", "ECDHE-ECDSA-AES128-GCM-SHA256")

	Http_SetHeader(w, "X-CSTP-License", "accept")
	Http_SetHeader(w, "X-CSTP-Routing-Filtering-Ignore", "false")
	Http_SetHeader(w, "X-CSTP-Quarantine", "false")
	Http_SetHeader(w, "X-CSTP-Disable-Always-On-VPN", "false")
	Http_SetHeader(w, "X-CSTP-Client-Bypass-Protocol", "false")
	Http_SetHeader(w, "X-CSTP-TCP-Keepalive", "false")
	// Http_SetHeader(w, "X-CSTP-Post-Auth-XML", ``)
	w.WriteHeader(http.StatusOK)

	hClone := w.Header().Clone()
	headers := make([]byte, 0)
	buf := bytes.NewBuffer(headers)
	_ = hClone.Write(buf)
	base.Debug(buf.String())

	hj := w.(http.Hijacker)
	conn, bufRW, err := hj.Hijack()
	if err != nil {
		base.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 开始数据处理
	switch base.Cfg.LinkMode {
	case base.LinkModeTUN:
		err = LinkTun(cSess)
	case base.LinkModeTAP:
		err = LinkTap(cSess)
	case base.LinkModeMacvtap:
		err = LinkMacvtap(cSess)
	}
	if err != nil {
		conn.Close()
		base.Error(err)
		return
	}

	go LinkCstp(conn, bufRW, cSess)
}
