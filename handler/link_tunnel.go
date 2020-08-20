package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bjdgyc/anylink/common"
)

var hn string

func init() {
	// 获取主机名称
	hn, _ = os.Hostname()
}

func LinkTunnel(w http.ResponseWriter, r *http.Request) {
	// TODO 调试信息输出
	// hd, _ := httputil.DumpRequest(r, true)
	// fmt.Println("DumpRequest: ", string(hd))
	fmt.Println("LinkTunnel", r.RemoteAddr)

	// 判断session-token的值
	cookie, err := r.Cookie("webvpn")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sess := SToken2Sess(cookie.Value)
	if sess == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 开启link
	cSess := sess.StartConn()
	if cSess == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 客户端信息
	cstp_mtu := r.Header.Get("X-CSTP-MTU")
	master_Secret := r.Header.Get("X-DTLS-Master-Secret")
	cSess.MasterSecret = master_Secret
	cSess.Mtu = cstp_mtu
	cSess.RemoteAddr = r.RemoteAddr

	w.Header().Set("Server", fmt.Sprintf("%s %s", common.APP_NAME, common.APP_VER))
	w.Header().Set("X-CSTP-Version", "1")
	w.Header().Set("X-CSTP-Protocol", "Copyright (c) 2004 Cisco Systems, Inc.")
	w.Header().Set("X-CSTP-Address", cSess.NetIp.String())    // 分配的ip地址
	w.Header().Set("X-CSTP-Netmask", common.ServerCfg.Ipv4Netmask) // 子网掩码
	w.Header().Set("X-CSTP-Hostname", hn)                          // 机器名称
	for _, v := range common.ServerCfg.ClientDns {
		w.Header().Add("X-CSTP-DNS", v) // dns地址
	}
	// 允许本地LAN访问vpn网络，必须放在路由的第一个
	if common.ServerCfg.AllowLan {
		w.Header().Set("X-CSTP-Split-Exclude", "0.0.0.0/255.255.255.255")
	}
	// 允许的路由
	for _, v := range common.ServerCfg.Include {
		w.Header().Add("X-CSTP-Split-Include", v)
	}
	// 不允许的路由
	for _, v := range common.ServerCfg.Exclude {
		w.Header().Add("X-CSTP-Split-Exclude", v)
	}
	// w.Header().Add("X-CSTP-Split-Include", "192.168.0.0/255.255.0.0")
	// w.Header().Add("X-CSTP-Split-Exclude", "10.1.5.2/255.255.255.255")

	w.Header().Set("X-CSTP-Lease-Duration", fmt.Sprintf("%d", common.IpLease)) // ip地址租期
	w.Header().Set("X-CSTP-Session-Timeout", "none")
	w.Header().Set("X-CSTP-Session-Timeout-Alert-Interval", "60")
	w.Header().Set("X-CSTP-Session-Timeout-Remaining", "none")
	w.Header().Set("X-CSTP-Idle-Timeout", "18000")
	w.Header().Set("X-CSTP-Disconnected-Timeout", "18000")
	w.Header().Set("X-CSTP-Keep", "true")
	w.Header().Set("X-CSTP-Tunnel-All-DNS", "false")
	w.Header().Set("X-CSTP-Rekey-Time", "5400")
	w.Header().Set("X-CSTP-Rekey-Method", "new-tunnel")
	w.Header().Set("X-CSTP-DPD", fmt.Sprintf("%d", common.ServerCfg.CstpDpd))             // 30 Dead peer detection in seconds
	w.Header().Set("X-CSTP-Keepalive", fmt.Sprintf("%d", common.ServerCfg.CstpKeepalive)) // 20
	w.Header().Set("X-CSTP-Banner", "welcome")                                            // urlencode
	w.Header().Set("X-CSTP-MSIE-Proxy-Lockdown", "true")
	w.Header().Set("X-CSTP-Smartcard-Removal-Disconnect", "true")

	w.Header().Set("X-CSTP-MTU", cstp_mtu) // 1399
	w.Header().Set("X-DTLS-MTU", cstp_mtu)

	w.Header().Set("X-DTLS-Session-ID", sess.DtlsSid)
	w.Header().Set("X-DTLS-Port", "4433")
	w.Header().Set("X-DTLS-Keepalive", fmt.Sprintf("%d", common.ServerCfg.CstpKeepalive))
	w.Header().Set("X-DTLS-Rekey-Time", "5400")
	w.Header().Set("X-DTLS12-CipherSuite", "ECDHE-ECDSA-AES128-GCM-SHA256")
	// w.Header().Set("X-DTLS12-CipherSuite", "ECDHE-RSA-AES128-GCM-SHA256")

	w.Header().Set("X-CSTP-License", "accept")
	w.Header().Set("X-CSTP-Routing-Filtering-Ignore", "false")
	w.Header().Set("X-CSTP-Quarantine", "false")
	w.Header().Set("X-CSTP-Disable-Always-On-VPN", "false")
	w.Header().Set("X-CSTP-Client-Bypass-Protocol", "false")
	w.Header().Set("X-CSTP-TCP-Keepalive", "false")
	// w.Header().Set("X-CSTP-Post-Auth-XML", ``)
	w.WriteHeader(http.StatusOK)

	hj := w.(http.Hijacker)
	conn, _, err := hj.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 开始数据处理
	go LinkTun(cSess)
	go LinkCstp(conn, cSess)
}
