package handler

import (
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/songgao/water/waterutil"
)

func payloadIn(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	// 进行Acl规则判断
	check := checkLinkAcl(cSess.Group, pl)
	if !check {
		// 校验不通过直接丢弃
		return false
	}

	closed := false
	select {
	case cSess.PayloadIn <- pl:
	case <-cSess.CloseChan:
		closed = true
	}

	return closed
}

func payloadOut(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	dSess := cSess.GetDtlsSession()
	if dSess == nil {
		return payloadOutCstp(cSess, pl)
	} else {
		return payloadOutDtls(cSess, dSess, pl)
	}
}

func payloadOutCstp(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	closed := false

	select {
	case cSess.PayloadOutCstp <- pl:
	case <-cSess.CloseChan:
		closed = true
	}

	return closed
}

func payloadOutDtls(cSess *sessdata.ConnSession, dSess *sessdata.DtlsSession, pl *sessdata.Payload) bool {
	select {
	case cSess.PayloadOutDtls <- pl:
	case <-dSess.CloseChan:
	}

	return false
}

// Acl规则校验
func checkLinkAcl(group *dbdata.Group, payload *sessdata.Payload) bool {
	if payload.LType == sessdata.LTypeIPData && payload.PType == 0x00 && len(group.LinkAcl) > 0 {
	} else {
		return true
	}

	data := *payload.Data
	ip_dst := waterutil.IPv4Destination(data)
	ip_port := waterutil.IPv4DestinationPort(data)
	ip_proto := waterutil.IPv4Protocol(data)
	// fmt.Println("sent:", ip_dst, ip_port)

	// 优先放行dns端口
	for _, v := range group.ClientDns {
		if v.Val == ip_dst.String() && ip_port == 53 {
			return true
		}
	}

	for _, v := range group.LinkAcl {
		// 循环判断ip和端口
		if v.IpNet.Contains(ip_dst) {
			// 放行允许ip的ping
			if v.Port == ip_port || v.Port == 0 || ip_proto == waterutil.ICMP {
				if v.Action == dbdata.Allow {
					return true
				} else {
					return false
				}
			}
		}
	}

	return false
}
