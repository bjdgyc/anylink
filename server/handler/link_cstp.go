package handler

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkCstp(conn net.Conn, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("LinkCstp return", cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		err     error
		n       int
		dataLen uint16
		dead    = time.Duration(cSess.CstpDpd+5) * time.Second
	)

	go cstpWrite(conn, cSess)

	for {

		// 设置超时限制
		err = conn.SetReadDeadline(time.Now().Add(dead))
		if err != nil {
			base.Error("SetDeadline: ", err)
			return
		}
		// hdata := make([]byte, BufferSize)
		hb := getByteFull()
		hdata := *hb
		n, err = conn.Read(hdata)
		if err != nil {
			base.Error("read hdata: ", err)
			return
		}

		// 限流设置
		err = cSess.RateLimit(n, true)
		if err != nil {
			base.Error(err)
		}

		switch hdata[6] {
		case 0x07: // KEEPALIVE
			// do nothing
			// base.Debug("recv keepalive", cSess.IpAddr)
		case 0x05: // DISCONNECT
			base.Debug("DISCONNECT", cSess.IpAddr)
			return
		case 0x03: // DPD-REQ
			// base.Debug("recv DPD-REQ", cSess.IpAddr)
			if payloadOutCstp(cSess, sessdata.LTypeIPData, 0x04, nil) {
				return
			}
		case 0x04:
			// log.Println("recv DPD-RESP")
		case 0x00: // DATA
			dataLen = binary.BigEndian.Uint16(hdata[4:6]) // 4,5
			if payloadIn(cSess, sessdata.LTypeIPData, 0x00, hdata[8:8+dataLen]) {
				return
			}
		}

		putByte(hb)
	}
}

func cstpWrite(conn net.Conn, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("cstpWrite return", cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		err error
		n   int
		// header  []byte
		payload *sessdata.Payload
	)

	for {
		select {
		case payload = <-cSess.PayloadOutCstp:
		case <-cSess.CloseChan:
			return
		}

		if payload.LType != sessdata.LTypeIPData {
			continue
		}

		h := []byte{'S', 'T', 'F', 0x01, 0x00, 0x00, payload.PType, 0x00}
		hb := getByteZero()
		header := *hb
		header = append(header, h...)
		if payload.PType == 0x00 {
			data := *payload.Data
			binary.BigEndian.PutUint16(header[4:6], uint16(len(data)))
			header = append(header, data...)
		}
		n, err = conn.Write(header)
		if err != nil {
			base.Error("write err", err)
			return
		}

		putByte(hb)
		putPayload(payload)

		// 限流设置
		err = cSess.RateLimit(n, false)
		if err != nil {
			base.Error(err)
		}
	}
}
