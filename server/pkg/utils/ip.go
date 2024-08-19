package utils

import (
	"encoding/binary"
	"net"
	"strings"
)

func Long2ip(i uint32) net.IP {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}

func Ip2long(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

// GetAddrIp 获取ip端口地址的ip数据
func GetAddrIp(s string) string {
	if strings.Contains(s, ":") {
		ss := s[:strings.LastIndex(s, ":")]
		if strings.HasPrefix(ss, "[") {
			return strings.Trim(ss, "[]")
		}
		return ss
	}

	return s
}
