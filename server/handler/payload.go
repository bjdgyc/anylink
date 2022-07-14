package handler

import (
	"encoding/binary"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/songgao/water/waterutil"
)

const (
	acc_proto_udp = iota + 1
	acc_proto_tcp
	acc_proto_https
	acc_proto_http
)

func payloadIn(cSess *sessdata.ConnSession, pl *sessdata.Payload) bool {
	if pl.LType == sessdata.LTypeIPData && pl.PType == 0x00 {
		// 进行Acl规则判断
		check := checkLinkAcl(cSess.Group, pl)
		if !check {
			// 校验不通过直接丢弃
			return false
		}

		logAudit(cSess, pl)
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
func checkLinkAcl(group *dbdata.Group, pl *sessdata.Payload) bool {
	if pl.LType == sessdata.LTypeIPData && pl.PType == 0x00 && len(group.LinkAcl) > 0 {
	} else {
		return true
	}

	ipDst := waterutil.IPv4Destination(pl.Data)
	ipPort := waterutil.IPv4DestinationPort(pl.Data)
	ipProto := waterutil.IPv4Protocol(pl.Data)
	// fmt.Println("sent:", ip_dst, ip_port)

	// 优先放行dns端口
	for _, v := range group.ClientDns {
		if v.Val == ipDst.String() && ipPort == 53 {
			return true
		}
	}

	for _, v := range group.LinkAcl {
		// 循环判断ip和端口
		if v.IpNet.Contains(ipDst) {
			// 放行允许ip的ping
			if v.Port == ipPort || v.Port == 0 || ipProto == waterutil.ICMP {
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

// 访问日志审计
func logAudit(cSess *sessdata.ConnSession, pl *sessdata.Payload) {
	if base.Cfg.AuditInterval < 0 {
		return
	}

	ipProto := waterutil.IPv4Protocol(pl.Data)
	// 访问协议
	var accessProto uint8 = acc_proto_tcp
	// 只统计 tcp和udp 的访问
	switch ipProto {
	case waterutil.TCP:
	case waterutil.UDP:
		accessProto = acc_proto_udp
	default:
		return
	}

	ipSrc := waterutil.IPv4Source(pl.Data)
	ipDst := waterutil.IPv4Destination(pl.Data)
	ipPort := waterutil.IPv4DestinationPort(pl.Data)

	b := getByte290()
	key := *b
	copy(key[:16], ipSrc)
	copy(key[16:32], ipDst)
	binary.BigEndian.PutUint16(key[32:34], ipPort)

	info := ""
	if ipProto == waterutil.TCP {
		accessProto, info = onTCP(waterutil.IPv4Payload(pl.Data))
	}
	key[34] = byte(accessProto)
	if info != "" {
		copy(key[35:35+len(info)], info)
	}

	s := utils.BytesToString(key)
	nu := utils.NowSec().Unix()

	// 判断已经存在，并且没有过期
	v, ok := cSess.IpAuditMap[s]
	if ok && nu-v < int64(base.Cfg.AuditInterval) {
		// 回收byte对象
		putByte290(b)
		return
	}

	cSess.IpAuditMap[s] = nu

	audit := dbdata.AccessAudit{
		Username:    cSess.Sess.Username,
		Protocol:    uint8(ipProto),
		Src:         ipSrc.String(),
		Dst:         ipDst.String(),
		DstPort:     ipPort,
		CreatedAt:   utils.NowSec(),
		AccessProto: accessProto,
		Info:        info,
	}

	_ = dbdata.Add(audit)
}
