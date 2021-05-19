package handler

import (
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkDtls(conn net.Conn, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("LinkDtls return", cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		dead = time.Duration(cSess.CstpDpd+5) * time.Second
	)

	go dtlsWrite(conn, cSess)

	for {
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
			base.Debug("recv DPD-REQ", cSess.IpAddr)
			if payloadOut(cSess, sessdata.LTypeIPData, 0x04, nil) {
				return
			}
		case 0x04:
			base.Debug("recv DPD-RESP", cSess.IpAddr)
		case 0x00: // DATA
			if payloadIn(cSess, sessdata.LTypeIPData, 0x00, hdata[1:]) {
				return
			}

		}
	}
}

func dtlsWrite(conn net.Conn, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("dtlsWrite return", cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		header  []byte
		payload *sessdata.Payload
	)

	for {
		select {
		case payload = <-cSess.PayloadOut:
		case <-cSess.CloseChan:
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
