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
	w.Header().Set("Server", fmt.Sprintf("%s %s", base.APP_NAME, base.APP_VER))
	w.Header().Set("X-CSTP-Version", "1")
	w.Header().Set("X-CSTP-Protocol", "Copyright (c) 2004 Cisco Systems, Inc.")
	w.Header().Set("X-CSTP-Address", cSess.IpAddr.String())             // 分配的ip地址
	w.Header().Set("X-CSTP-Netmask", sessdata.IpPool.Ipv4Mask.String()) // 子网掩码
	w.Header().Set("X-CSTP-Hostname", hn)                               // 机器名称

	// 允许本地LAN访问vpn网络，必须放在路由的第一个
	if cSess.Group.AllowLan {
		w.Header().Set("X-CSTP-Split-Exclude", "0.0.0.0/255.255.255.255")
	}
	// dns地址
	for _, v := range cSess.Group.ClientDns {
		w.Header().Add("X-CSTP-DNS", v.Val)
	}
	// 允许的路由
	for _, v := range cSess.Group.RouteInclude {
		w.Header().Add("X-CSTP-Split-Include", v.IpMask)
	}
	// 不允许的路由
	for _, v := range cSess.Group.RouteExclude {
		w.Header().Add("X-CSTP-Split-Exclude", v.IpMask)
	}

	w.Header().Set("X-CSTP-Lease-Duration", fmt.Sprintf("%d", base.Cfg.IpLease)) // ip地址租期
	w.Header().Set("X-CSTP-Session-Timeout", "none")
	w.Header().Set("X-CSTP-Session-Timeout-Alert-Interval", "60")
	w.Header().Set("X-CSTP-Session-Timeout-Remaining", "none")
	w.Header().Set("X-CSTP-Idle-Timeout", "18000")
	w.Header().Set("X-CSTP-Disconnected-Timeout", "18000")
	w.Header().Set("X-CSTP-Keep", "true")
	w.Header().Set("X-CSTP-Tunnel-All-DNS", "false")

	w.Header().Set("X-CSTP-Rekey-Time", "172800")
	w.Header().Set("X-CSTP-Rekey-Method", "new-tunnel")

	w.Header().Set("X-CSTP-DPD", fmt.Sprintf("%d", cstpDpd))
	w.Header().Set("X-CSTP-Keepalive", fmt.Sprintf("%d", cstpKeepalive))
	// w.Header().Set("X-CSTP-Banner", banner.Banner)
	w.Header().Set("X-CSTP-MSIE-Proxy-Lockdown", "true")
	w.Header().Set("X-CSTP-Smartcard-Removal-Disconnect", "true")

	w.Header().Set("X-CSTP-MTU", fmt.Sprintf("%d", cSess.Mtu)) // 1399
	w.Header().Set("X-DTLS-MTU", fmt.Sprintf("%d", cSess.Mtu))

	w.Header().Set("X-DTLS-Session-ID", sess.DtlsSid)
	w.Header().Set("X-DTLS-Port", dtlsPort)
	w.Header().Set("X-DTLS-DPD", fmt.Sprintf("%d", cstpDpd))
	w.Header().Set("X-DTLS-Keepalive", fmt.Sprintf("%d", cstpKeepalive))
	w.Header().Set("X-DTLS-Rekey-Time", "5400")
	w.Header().Set("X-DTLS12-CipherSuite", "ECDHE-ECDSA-AES128-GCM-SHA256")

	w.Header().Set("X-CSTP-License", "accept")
	w.Header().Set("X-CSTP-Routing-Filtering-Ignore", "false")
	w.Header().Set("X-CSTP-Quarantine", "false")
	w.Header().Set("X-CSTP-Disable-Always-On-VPN", "false")
	w.Header().Set("X-CSTP-Client-Bypass-Protocol", "false")
	w.Header().Set("X-CSTP-TCP-Keepalive", "false")
	// w.Header().Set("X-CSTP-Post-Auth-XML", ``)
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
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go LinkCstp(conn, bufRW, cSess)
}
