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

		switch pl.Data[6] {
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
			// 获取数据长度
			dataLen = binary.BigEndian.Uint16(pl.Data[4:6]) // 4,5
			// 去除数据头
			copy(pl.Data, pl.Data[8:8+dataLen])
			// 更新切片长度
			pl.Data = pl.Data[:dataLen]
			// pl.Data = append(pl.Data[:0], pl.Data[8:8+dataLen]...)
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
		err error
		n   int
		pl  *sessdata.Payload
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

		if pl.PType == 0x00 {
			// 获取数据长度
			l := len(pl.Data)
			// 先扩容 +8
			pl.Data = pl.Data[:l+8]
			// 数据后移
			copy(pl.Data[8:], pl.Data)
			// 添加头信息
			copy(pl.Data[:8], plHeader)
			// 更新头长度
			binary.BigEndian.PutUint16(pl.Data[4:6], uint16(l))
		} else {
			pl.Data = append(pl.Data[:0], plHeader...)
			// 设置头类型
			pl.Data[6] = pl.PType
		}

		n, err = conn.Write(pl.Data)
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
