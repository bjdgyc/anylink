package handler

import "github.com/bjdgyc/anylink/sessdata"

func payloadIn(cSess *sessdata.ConnSession, lType sessdata.LType, pType byte, data []byte) bool {
	payload := &sessdata.Payload{
		LType: lType,
		PType: pType,
		Data:  data,
	}

	return payloadInData(cSess, payload)
}

func payloadInData(cSess *sessdata.ConnSession, payload *sessdata.Payload) bool {
	closed := false

	select {
	case cSess.PayloadIn <- payload:
	case <-cSess.CloseChan:
		closed = true
	}

	return closed
}

func payloadOut(cSess *sessdata.ConnSession, lType sessdata.LType, pType byte, data []byte) bool {
	payload := &sessdata.Payload{
		LType: lType,
		PType: pType,
		Data:  data,
	}

	return payloadOutData(cSess, payload)
}

func payloadOutData(cSess *sessdata.ConnSession, payload *sessdata.Payload) bool {
	closed := false

	select {
	case cSess.PayloadOut <- payload:
	case <-cSess.CloseChan:
		closed = true
	}

	return closed
}
