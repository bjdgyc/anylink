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
		data    []byte
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
		pl := getPayload()
		data = *pl.Data
		n, err = conn.Read(data)
		if err != nil {
			base.Error("read hdata: ", err)
			return
		}

		// 限流设置
		err = cSess.RateLimit(n, true)
		if err != nil {
			base.Error(err)
		}

		switch data[6] {
		case 0x07: // KEEPALIVE
			// do nothing
			// base.Debug("recv keepalive", cSess.IpAddr)
		case 0x05: // DISCONNECT
			base.Debug("DISCONNECT", cSess.IpAddr)
			return
		case 0x03: // DPD-REQ
			// base.Debug("recv DPD-REQ", cSess.IpAddr)
			pl.PType = 0x04
			if payloadOutCstp(cSess, pl) {
				return
			}
		case 0x04:
			// log.Println("recv DPD-RESP")
		case 0x00: // DATA
			dataLen = binary.BigEndian.Uint16(data[4:6]) // 4,5
			copy(data, data[8:8+dataLen])
			*pl.Data = data[:dataLen]
			if payloadIn(cSess, pl) {
				return
			}
		}
	}
}

func cstpWrite(conn net.Conn, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("cstpWrite return", cSess.IpAddr)
		_ = conn.Close()
		cSess.Close()
	}()

	var (
		err  error
		n    int
		data []byte
		pl   *sessdata.Payload
	)

	for {
		select {
		case pl = <-cSess.PayloadOutCstp:
		case <-cSess.CloseChan:
			return
		}

		if pl.LType != sessdata.LTypeIPData {
			continue
		}

		data = *pl.Data
		if pl.PType == 0x00 {
			l := len(data)
			data = data[:l+8]
			copy(data[8:], data)
			copy(data[:8], plHeader)
			binary.BigEndian.PutUint16(data[4:6], uint16(l))
		} else {
			data = append(data[:0], plHeader...)
			data[6] = pl.PType
		}
		*pl.Data = data
		n, err = conn.Write(*pl.Data)
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
