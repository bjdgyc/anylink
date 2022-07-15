package handler

import (
	"bufio"
	"bytes"
	"net/http"
	"regexp"
)

var tcpParsers = []func([]byte) (uint8, string){
	sniParser,
	httpParser,
}

var (
	sniRe = regexp.MustCompile("\x00\x00.{4}\x00.{2}([a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,6})\x00")
)

func onTCP(payload []byte) (uint8, string) {
	ihl := (payload[12] & 0xf0) >> 2
	data := payload[ihl:]
	for _, parser := range tcpParsers {
		if proto, info := parser(data); info != "" {
			return proto, info
		}
	}
	return acc_proto_tcp, ""
}

func sniParser(data []byte) (uint8, string) {
	dataSize := len(data)
	if dataSize < 2 || data[0] != 0x16 || data[1] != 0x03 {
		return acc_proto_tcp, ""
	}
	m := sniRe.FindSubmatch(data)
	if len(m) < 2 {
		return acc_proto_tcp, ""
	}
	host := string(m[1])
	return acc_proto_https, host
}

func httpParser(data []byte) (uint8, string) {
	if req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data))); err == nil {
		return acc_proto_http, req.Host
	}
	return acc_proto_tcp, ""
}
