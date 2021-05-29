package handler

import (
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkDtls(conn net.Conn, cSess *sessdata.ConnSession) {
	base.Debug("LinkDtls connect ip:", cSess.IpAddr, "udp-rip:", conn.RemoteAddr())
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
		dead = time.Duration(cSess.CstpDpd+5) * time.Second
	)

	go dtlsWrite(conn, dSess, cSess)

	for {
		err := conn.SetReadDeadline(time.Now().Add(dead))
		if err != nil {
			base.Error("SetDeadline: ", err)
			return
		}

		// hdata := make([]byte, BufferSize)
		hdata := getByteFull()
		n, err := conn.Read(hdata)
		if err != nil {
			base.Error("read hdata: ", err)
			return
		}

		// 限流设置
		err = cSess.RateLimit(n, true)
		if err != nil {
			base.Error(err)
		}

		switch hdata[0] {
		case 0x07: // KEEPALIVE
			// do nothing
			// base.Debug("recv keepalive", cSess.IpAddr)
		case 0x05: // DISCONNECT
			base.Debug("DISCONNECT DTLS", cSess.IpAddr)
			return
		case 0x03: // DPD-REQ
			// base.Debug("recv DPD-REQ", cSess.IpAddr)
			if payloadOutDtls(cSess, dSess, sessdata.LTypeIPData, 0x04, nil) {
				return
			}
		case 0x04:
			// base.Debug("recv DPD-RESP", cSess.IpAddr)
		case 0x00: // DATA
			if payloadIn(cSess, sessdata.LTypeIPData, 0x00, hdata[1:n]) {
				return
			}
		}

		putByte(hdata)
	}
}

func dtlsWrite(conn net.Conn, dSess *sessdata.DtlsSession, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("dtlsWrite return", cSess.IpAddr)
		_ = conn.Close()
		dSess.Close()
	}()

	var (
		// header  []byte
		payload *sessdata.Payload
	)

	for {
		// dtls优先推送数据
		select {
		case payload = <-cSess.PayloadOutDtls:
		case <-dSess.CloseChan:
			return
		}

		if payload.LType != sessdata.LTypeIPData {
			continue
		}

		// header = []byte{payload.PType}
		header := getByteZero()
		header = append(header, payload.PType)
		if payload.PType == 0x00 { // data
			header = append(header, payload.Data...)
		}
		n, err := conn.Write(header)
		if err != nil {
			base.Error("write err", err)
			return
		}

		putByte(header)
		putPayload(payload)

		// 限流设置
		err = cSess.RateLimit(n, false)
		if err != nil {
			base.Error(err)
		}
	}
}
