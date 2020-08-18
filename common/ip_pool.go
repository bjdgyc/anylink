package common

import (
	"encoding/binary"
	"math"
	"net"
	"sync"
	"time"
)

const (
	// ip租期 (秒)
	IpLease = 1209600
)

var (
	ipPool = &IpPoolConfig{}
	macIps = map[string]*MacIp{}
)

type MacIp struct {
	IsActive  bool
	Ip        net.IP
	MacAddr   string
	LastLogin time.Time
}

type IpPoolConfig struct {
	mux sync.Mutex
	// 计算动态ip
	Ipv4Net     *net.IPNet
	Ipv4GateWay net.IP
	IpLongMin   uint32
	IpLongMax   uint32
	IpLongNow   uint32
}

func initIpPool() {
	// ip地址
	ip := net.ParseIP(ServerCfg.Ipv4Network)
	// 子网掩码
	maskIp := net.ParseIP(ServerCfg.Ipv4Netmask).To4()
	mask := net.IPMask(maskIp)

	ipNet := &net.IPNet{IP: ip, Mask: mask}
	ipPool.Ipv4Net = ipNet

	// 网络地址零值
	min := binary.BigEndian.Uint32(ip.Mask(mask))
	// 广播地址
	one, _ := ipNet.Mask.Size()
	max := min | uint32(math.Pow(2, float64(32-one))-1)

	min += 1 // 网关
	ipPool.Ipv4GateWay = long2ip(min)
	ServerCfg.Ipv4GateWay = ipPool.Ipv4GateWay.String()
	// 第一个可用地址
	min += 1
	ipPool.IpLongMin = min
	ipPool.IpLongMax = max
	ipPool.IpLongNow = min
}

func long2ip(i uint32) net.IP {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}

// 获取动态ip
func AcquireIp(macAddr string) net.IP {
	ipPool.mux.Lock()
	defer ipPool.mux.Unlock()
	tNow := time.Now()

	// 判断已经分配过
	if mi, ok := macIps[macAddr]; ok {
		mi.IsActive = true
		mi.LastLogin = tNow
		return mi.Ip
	}

	// ip池分配完之前
	if ipPool.IpLongNow < ipPool.IpLongMax {
		// 递增分配一个ip
		ip := long2ip(ipPool.IpLongNow)
		mi := &MacIp{IsActive: true, Ip: ip, MacAddr: macAddr, LastLogin: tNow}
		macIps[macAddr] = mi
		ipPool.IpLongNow += 1
		return ip
	}

	// 查找过期数据
	farMi := &MacIp{LastLogin: tNow}
	for k, v := range macIps {
		// 跳过活跃连接
		if v.IsActive {
			continue
		}

		// 已经超过租期
		if tNow.Sub(v.LastLogin) > IpLease*time.Second {
			delete(macIps, k)
			ip := v.Ip
			mi := &MacIp{IsActive: true, Ip: ip, MacAddr: macAddr, LastLogin: tNow}
			macIps[macAddr] = mi
			return ip
		}

		// 其他情况判断最早登陆的mac
		if v.LastLogin.Before(farMi.LastLogin) {
			farMi = v
		}
	}

	// 全都在线，没有数据可用
	if farMi.MacAddr == "" {
		return nil
	}

	// 使用最早登陆的mac地址
	delete(macIps, farMi.MacAddr)
	ip := farMi.Ip
	mi := &MacIp{IsActive: true, Ip: ip, MacAddr: macAddr, LastLogin: tNow}
	macIps[macAddr] = mi
	return ip
}

// 回收ip
func ReleaseIp(ip net.IP, macAddr string) {
	ipPool.mux.Lock()
	defer ipPool.mux.Unlock()
	if mi, ok := macIps[macAddr]; ok {
		if mi.Ip.Equal(ip) {
			mi.IsActive = false
			mi.LastLogin = time.Now()
		}
	}
}
