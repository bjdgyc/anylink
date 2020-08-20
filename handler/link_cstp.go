package handler

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/bjdgyc/anylink/common"
)

func LinkCstp(conn net.Conn, sess *ConnSession) {
	// fmt.Println("HandlerCstp")
	defer func() {
		log.Println("LinkCstp return")
		conn.Close()
		sess.Close()
	}()

	var (
		err     error
		dataLen uint16
		dead    = time.Duration(common.ServerCfg.CstpDpd+2) * time.Second
	)

	go cstpWrite(conn, sess)

	for {
		// 设置超时限制
		err = conn.SetDeadline(time.Now().Add(dead))
		if err != nil {
			log.Println("SetDeadline: ", err)
			return
		}
		hdata := make([]byte, 1500)
		_, err = conn.Read(hdata)
		if err != nil {
			log.Println("read hdata: ", err)
			return
		}

		switch hdata[6] {
		case 0x07: // KEEPALIVE
			// do nothing
			// fmt.Println("keepalive")
		case 0x05: // DISCONNECT
			// fmt.Println("DISCONNECT")
			return
		case 0x03: // DPD-REQ
			fmt.Println("DPD-REQ")
			payload := &Payload{
				ptype: 0x04, // DPD-RESP
			}
			// 直接返回给客户端 resp
			select {
			case sess.PayloadOut <- payload:
			case <-sess.Closed:
				return
			}
			break
		case 0x00:
			dataLen = binary.BigEndian.Uint16(hdata[4:6]) // 4,5
			payload := &Payload{
				ptype: 0x00, // DPD-RESP
				data:  hdata[8 : 8+dataLen],
			}
			select {
			case sess.PayloadIn <- payload:
			case <-sess.Closed:
				return
			}
		}
	}
}

func cstpWrite(conn net.Conn, sess *ConnSession) {
	defer func() {
		log.Println("cstpWrite return")
		conn.Close()
		sess.Close()
	}()

	var (
		err     error
		header  []byte
		payload *Payload
	)

	for {
		select {
		case payload = <-sess.PayloadOut:
		case <-sess.Closed:
			return
		}

		header = []byte{'S', 'T', 'F', 0x01, 0x00, 0x00, payload.ptype, 0x00}
		if payload.ptype == 0x00 { // data
			binary.BigEndian.PutUint16(header[4:6], uint16(len(payload.data)))
			header = append(header, payload.data...)
		}
		_, err = conn.Write(header)
		if err != nil {
			log.Println("write err", err)
			return
		}
	}
}
