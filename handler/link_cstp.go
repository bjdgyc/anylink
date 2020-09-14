package handler

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/bjdgyc/anylink/common"
	"github.com/bjdgyc/anylink/sessdata"
)

func LinkCstp(conn net.Conn, sess *sessdata.ConnSession) {
	log.Println("HandlerCstp")
	sessdata.Sess = sess
	defer func() {
		log.Println("LinkCstp return")
		conn.Close()
		sess.Close()
	}()

	var (
		err     error
		n       int
		dataLen uint16
		dead    = time.Duration(common.ServerCfg.CstpDpd+2) * time.Second
	)

	go cstpWrite(conn, sess)

	for {

		// 设置超时限制
		err = conn.SetReadDeadline(time.Now().Add(dead))
		if err != nil {
			log.Println("SetDeadline: ", err)
			return
		}
		hdata := make([]byte, BufferSize)
		n, err = conn.Read(hdata)
		if err != nil {
			log.Println("read hdata: ", err)
			return
		}

		// 限流设置
		err = sess.RateLimit(n, true)
		if err != nil {
			log.Println(err)
		}

		switch hdata[6] {
		case 0x07: // KEEPALIVE
			// do nothing
			// log.Println("recv keepalive")
		case 0x05: // DISCONNECT
			// log.Println("DISCONNECT")
			return
		case 0x03: // DPD-REQ
			// log.Println("recv DPD-REQ")
			if payloadOut(sess, sessdata.LTypeIPData, 0x04, nil) {
				return
			}
		case 0x04:
			// log.Println("recv DPD-RESP")
		case 0x00: // DATA
			dataLen = binary.BigEndian.Uint16(hdata[4:6]) // 4,5
			data := hdata[8 : 8+dataLen]

			if payloadIn(sess, sessdata.LTypeIPData, 0x00, data) {
				return
			}

		}
	}
}

func cstpWrite(conn net.Conn, sess *sessdata.ConnSession) {
	defer func() {
		log.Println("cstpWrite return")
		conn.Close()
		sess.Close()
	}()

	var (
		err     error
		n       int
		header  []byte
		payload *sessdata.Payload
	)

	for {
		select {
		case payload = <-sess.PayloadOut:
		case <-sess.CloseChan:
			return
		}

		if payload.LType != sessdata.LTypeIPData {
			continue
		}

		header = []byte{'S', 'T', 'F', 0x01, 0x00, 0x00, payload.PType, 0x00}
		if payload.PType == 0x00 { // data
			binary.BigEndian.PutUint16(header[4:6], uint16(len(payload.Data)))
			header = append(header, payload.Data...)
		}
		n, err = conn.Write(header)
		if err != nil {
			log.Println("write err", err)
			return
		}

		// 限流设置
		err = sess.RateLimit(n, false)
		if err != nil {
			log.Println(err)
		}
	}
}
