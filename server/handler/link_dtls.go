package handler

import (
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkDtls(conn net.Conn, cSess *sessdata.ConnSession) {
	base.Debug("LinkDtls connect ip:", cSess.IpAddr, "user:", cSess.Username, "udp-rip:", conn.RemoteAddr())
	dSess := cSess.NewDtlsConn()
	if dSess == nil {
		// 创建失败，直接关闭链接
		_ = conn.Close()
		return
	}

	defer func() {
		base.Debug("LinkDtls return", cSess.IpAddr)
		_ = conn.Close()
		dSess.Close()
	}()

	var (
		err  error
		n    int
		dead = time.Duration(cSess.CstpDpd+5) * time.Second
	)

	go dtlsWrite(conn, dSess, cSess)

	for {
		err = conn.SetReadDeadline(utils.NowSec().Add(dead))
		if err != nil {
			base.Error("SetDeadline: ", err)
			return
		}

		pl := getPayload()
		n, err = conn.Read(pl.Data)
		if err != nil {
			base.Error("read hdata: ", err)
			return
		}

		// 限流设置
		err = cSess.RateLimit(n, true)
		if err != nil {
			base.Error(err)
		}

		switch pl.Data[0] {
		case 0x07: // KEEPALIVE
			// do nothing
			// base.Debug("recv keepalive", cSess.IpAddr)
		case 0x05: // DISCONNECT
			base.Debug("DISCONNECT DTLS", cSess.IpAddr)
			return
		case 0x03: // DPD-REQ
			// base.Debug("recv DPD-REQ", cSess.IpAddr)
			pl.PType = 0x04
			if payloadOutDtls(cSess, dSess, pl) {
				return
			}
		case 0x04:
			// base.Debug("recv DPD-RESP", cSess.IpAddr)
		case 0x00: // DATA
			// 去除数据头
			// copy(pl.Data, pl.Data[1:n])
			// 更新切片长度
			// pl.Data = pl.Data[:n-1]
			pl.Data = append(pl.Data[:0], pl.Data[1:n]...)
			if payloadIn(cSess, pl) {
				return
			}
		}

	}
}

func dtlsWrite(conn net.Conn, dSess *sessdata.DtlsSession, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("dtlsWrite return", cSess.IpAddr)
		_ = conn.Close()
		dSess.Close()
	}()

	var (
		pl *sessdata.Payload
	)

	for {
		// dtls优先推送数据
		select {
		case pl = <-cSess.PayloadOutDtls:
		case <-dSess.CloseChan:
			return
		}

		if pl.LType != sessdata.LTypeIPData {
			continue
		}

		// header = []byte{payload.PType}
		if pl.PType == 0x00 { // data
			// 获取数据长度
			l := len(pl.Data)
			// 先扩容 +1
			pl.Data = pl.Data[:l+1]
			// 数据后移
			copy(pl.Data[1:], pl.Data)
			// 添加头信息
			pl.Data[0] = pl.PType
		} else {
			// 设置头类型
			pl.Data = append(pl.Data[:0], pl.PType)
		}
		n, err := conn.Write(pl.Data)
		if err != nil {
			base.Error("write err", err)
			return
		}

		putPayload(pl)

		// 限流设置
		err = cSess.RateLimit(n, false)
		if err != nil {
			base.Error(err)
		}
	}
}
