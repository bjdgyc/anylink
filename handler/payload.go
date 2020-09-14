package handler

import "github.com/bjdgyc/anylink/sessdata"

func payloadIn(sess *sessdata.ConnSession, lType sessdata.LType, pType byte, data []byte) bool {
	payload := &sessdata.Payload{
		LType: lType,
		PType: pType,
		Data:  data,
	}

	return payloadInData(sess, payload)
}

func payloadInData(sess *sessdata.ConnSession, payload *sessdata.Payload) bool {
	closed := false

	select {
	case sess.PayloadIn <- payload:
	case <-sess.CloseChan:
		closed = true
	}

	return closed
}

func payloadOut(sess *sessdata.ConnSession, lType sessdata.LType, pType byte, data []byte) bool {
	payload := &sessdata.Payload{
		LType: lType,
		PType: pType,
		Data:  data,
	}

	return payloadOutData(sess, payload)
}

func payloadOutData(sess *sessdata.ConnSession, payload *sessdata.Payload) bool {
	closed := false

	select {
	case sess.PayloadOut <- payload:
	case <-sess.CloseChan:
		closed = true
	}

	return closed
}
