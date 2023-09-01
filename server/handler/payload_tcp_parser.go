package handler

import (
	"bufio"
	"bytes"
	"net/http"
	"regexp"
	"strings"
)

var tcpParsers = []func([]byte) (uint8, string){
	sniNewParser,
	httpParser,
}

func onTCP(payload []byte) (uint8, string) {
	size := len(payload)
	ihl := (payload[12] & 0xf0) >> 2
	if int(ihl) > size {
		return acc_proto_tcp, ""
	}
	data := payload[ihl:]
	for _, parser := range tcpParsers {
		if proto, info := parser(data); proto != acc_proto_tcp {
			return proto, info
		}
	}
	return acc_proto_tcp, ""
}

func sniNewParser(b []byte) (uint8, string) {
	if len(b) < 2 || b[0] != 0x16 || b[1] != 0x03 {
		return acc_proto_tcp, ""
	}
	if len(b) < 6 {
		return acc_proto_tcp, ""
	}	
	rest := b[5:]
	restLen := len(rest)
	if restLen == 0 {
		return acc_proto_tcp, ""
	}
	current := 0
	handshakeType := rest[0]
	current += 1
	if handshakeType != 0x1 {
		return acc_proto_tcp, ""
	}
	// Skip over another length
	current += 3
	// Skip over protocolversion
	current += 2
	// Skip over random number
	current += 4 + 28
	if current >= restLen {
		return acc_proto_https, ""
	}
	// Skip over session ID
	sessionIDLength := int(rest[current])
	current += 1
	current += sessionIDLength
	if current+1 >= restLen {
		return acc_proto_https, ""
	}
	cipherSuiteLength := (int(rest[current]) << 8) + int(rest[current+1])
	current += 2
	current += cipherSuiteLength
	if current >= restLen {
		return acc_proto_https, ""
	}
	compressionMethodLength := int(rest[current])
	current += 1
	current += compressionMethodLength

	if current >= restLen {
		return acc_proto_https, ""
	}
	current += 2
	hostname := ""
	for current+4 < restLen && hostname == "" {
		extensionType := (int(rest[current]) << 8) + int(rest[current+1])
		current += 2
		extensionDataLength := (int(rest[current]) << 8) + int(rest[current+1])
		current += 2
		if extensionType == 0 {
			// Skip over number of names as we're assuming there's just one
			current += 2
			if current >= restLen {
				return acc_proto_https, ""
			}
			nameType := rest[current]
			current += 1
			if nameType != 0 {
				return acc_proto_https, ""
			}
			if current+1 >= restLen {
				return acc_proto_https, ""
			}
			nameLen := (int(rest[current]) << 8) + int(rest[current+1])
			current += 2
			if current+nameLen >= restLen {
				return acc_proto_https, ""
			}
			hostname = string(rest[current : current+nameLen])
		}
		current += extensionDataLength
	}
	if hostname == "" {
		return acc_proto_https, ""
	}
	if !validDomainChar(hostname) {
		return acc_proto_https, ""
	}
	return acc_proto_https, hostname
}

// Beta
func httpNewParser(data []byte) (uint8, string) {
	methodArr := []string{"OPTIONS", "HEAD", "GET", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}
	pos := bytes.IndexByte(data, 10)
	if pos == -1 {
		return acc_proto_tcp, ""
	}
	method, uri, _ := strings.Cut(string(data[:pos]), " ")
	ok := false
	for _, v := range methodArr {
		if v == method {
			ok = true
		}
	}
	if !ok {
		return acc_proto_tcp, ""
	}
	hostname := ""
	// GET http://www.google.com/index.html HTTP/1.1
	if len(uri) > 7 && uri[:4] == "http" {
		uriSlice := strings.Split(uri[7:], "/")
		hostname = uriSlice[0]
		return acc_proto_http, hostname
	}
	packet := string(data)
	hostPos := strings.Index(packet, "Host: ")
	if hostPos == -1 {
		hostPos = strings.Index(packet, "HOST: ")
		if hostPos == -1 {
			return acc_proto_tcp, ""
		}
	}
	hostEndPos := strings.Index(packet[hostPos:], "\n")
	if hostEndPos == -1 {
		return acc_proto_tcp, ""
	}
	hostname = packet[hostPos+6 : hostPos+hostEndPos-1]
	return acc_proto_http, hostname
}

func sniParser(data []byte) (uint8, string) {
	if len(data) < 2 || data[0] != 0x16 || data[1] != 0x03 {
		return acc_proto_tcp, ""
	}
	sniRe := regexp.MustCompile("\x00\x00.{4}\x00.{2}([a-z0-9]+([\\-\\.]{1}[a-z0-9]+)*\\.[a-z]{2,6})\x00")
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

// 校验域名的合法字符, 处理乱码问题
func validDomainChar(addr string) bool {
	// Allow a-z A-Z . - 0-9
	for i := 0; i < len(addr); i++ {
		c := addr[i]
		if !((c >= 97 && c <= 122) || (c >= 65 && c <= 90) || (c >= 45 && c <= 46) || (c >= 48 && c <= 57)) {
			return false
		}
	}
	return true
}
