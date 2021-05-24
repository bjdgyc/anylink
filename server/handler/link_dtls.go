package handler

import (
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkDtls(conn net.Conn, cSess *sessdata.ConnSession) {
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

	now := time.Now()

	for {

		if time.Now().Sub(now) > time.Second*30 {
			// return
		}

		err := conn.SetReadDeadline(time.Now().Add(dead))
		if err != nil {
			base.Error("SetDeadline: ", err)
			return
		}
		hdata := make([]byte, BufferSize)
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
			base.Debug("recv keepalive", cSess.IpAddr)
		case 0x05: // DISCONNECT
			base.Debug("DISCONNECT", cSess.IpAddr)
			return
		case 0x03: // DPD-REQ
			// base.Debug("recv DPD-REQ", cSess.IpAddr)
			payload := &sessdata.Payload{
				LType: sessdata.LTypeIPData,
				PType: 0x04,
				Data:  nil,
			}

			select {
			case cSess.PayloadOutDtls <- payload:
			case <-dSess.CloseChan:
				return
			}
		case 0x04:
			// base.Debug("recv DPD-RESP", cSess.IpAddr)
		case 0x00: // DATA
			if payloadIn(cSess, sessdata.LTypeIPData, 0x00, hdata[1:n]) {
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
		header  []byte
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

		header = []byte{payload.PType}
		header = append(header, payload.Data...)
		n, err := conn.Write(header)
		if err != nil {
			base.Error("write err", err)
			return
		}

		// 限流设置
		err = cSess.RateLimit(n, false)
		if err != nil {
			base.Error(err)
		}
	}
}
