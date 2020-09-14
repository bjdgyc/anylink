package sessdata

import (
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/common"
	"github.com/bjdgyc/anylink/dbdata"
)

const (
	// ip租期 (秒)
	IpLease = 1209600
)

var (
	IpPool  = &IpPoolConfig{}
	macInfo = map[string]*MacIp{}
	ipInfo  = map[string]*MacIp{}
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
	Ipv4Gateway net.IP
	Ipv4IPNet   net.IPNet
	IpLongMin   uint32
	IpLongMax   uint32
}

func initIpMac() {
	macs := dbdata.GetAllMacIp()
	for _, v := range macs {
		mi := &MacIp{}
		CopyStruct(mi, v)
		macInfo[v.MacAddr] = mi
		ipInfo[v.Ip.String()] = mi
	}
}

func initIpPool() {

	// 地址处理
	// ip地址
	ip := net.ParseIP(common.ServerCfg.Ipv4Network)
	// 子网掩码
	maskIp := net.ParseIP(common.ServerCfg.Ipv4Netmask).To4()
	IpPool.Ipv4IPNet = net.IPNet{IP: ip, Mask: net.IPMask(maskIp)}
	IpPool.Ipv4Gateway = net.ParseIP(common.ServerCfg.Ipv4Gateway)

	// 网络地址零值
	// zero := binary.BigEndian.Uint32(ip.Mask(mask))
	// 广播地址
	// one, _ := ipNet.Mask.Size()
	// max := min | uint32(math.Pow(2, float64(32-one))-1)

	// ip地址池
	IpPool.IpLongMin = ip2long(net.ParseIP(common.ServerCfg.Ipv4Pool[0]))
	IpPool.IpLongMax = ip2long(net.ParseIP(common.ServerCfg.Ipv4Pool[1]))
}

func long2ip(i uint32) net.IP {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, i)
	return ip
}

func ip2long(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

// 获取动态ip
func AcquireIp(macAddr string) net.IP {
	IpPool.mux.Lock()
	defer IpPool.mux.Unlock()
	tNow := time.Now()

	// 判断已经分配过
	if mi, ok := macInfo[macAddr]; ok {
		ip := mi.Ip
		// 检测原有ip是否在新的ip池内
		if IpPool.Ipv4IPNet.Contains(ip) {
			mi.IsActive = true
			mi.LastLogin = tNow
			// 回写db数据
			dbdata.Set(dbdata.BucketMacIp, macAddr, mi)
			return ip
		} else {
			delete(macInfo, macAddr)
			delete(ipInfo, ip.String())
			dbdata.Del(dbdata.BucketMacIp, macAddr)
		}
	}

	farMac := &MacIp{LastLogin: tNow}
	// 全局遍历未分配ip
	for i := IpPool.IpLongMin; i <= IpPool.IpLongMax; i++ {
		ip := long2ip(i)
		ipStr := ip.String()
		v, ok := ipInfo[ipStr]
		// 该ip没有被使用
		if !ok {
			mi := &MacIp{IsActive: true, Ip: ip, MacAddr: macAddr, LastLogin: tNow}
			macInfo[macAddr] = mi
			ipInfo[ipStr] = mi
			// 回写db数据
			dbdata.Set(dbdata.BucketMacIp, macAddr, mi)
			return ip
		}

		// 跳过活跃连接
		if v.IsActive {
			continue
		}
		// 已经超过租期
		if tNow.Sub(v.LastLogin) > IpLease*time.Second {
			delete(macInfo, v.MacAddr)
			mi := &MacIp{IsActive: true, Ip: ip, MacAddr: macAddr, LastLogin: tNow}
			macInfo[macAddr] = mi
			ipInfo[ipStr] = mi
			// 回写db数据
			dbdata.Del(dbdata.BucketMacIp, v.MacAddr)
			dbdata.Set(dbdata.BucketMacIp, macAddr, mi)
			return ip
		}
		// 其他情况判断最早登陆的mac
		if v.LastLogin.Before(farMac.LastLogin) {
			farMac = v
		}
	}

	// 全都在线，没有数据可用
	if farMac.MacAddr == "" {
		return nil
	}

	// 使用最早登陆的mac ip
	delete(macInfo, farMac.MacAddr)
	ip := farMac.Ip
	mi := &MacIp{IsActive: true, Ip: ip, MacAddr: macAddr, LastLogin: tNow}
	macInfo[macAddr] = mi
	ipInfo[ip.String()] = mi
	// 回写db数据
	dbdata.Del(dbdata.BucketMacIp, farMac.MacAddr)
	dbdata.Set(dbdata.BucketMacIp, macAddr, mi)
	return ip
}

// 回收ip
func ReleaseIp(ip net.IP, macAddr string) {
	IpPool.mux.Lock()
	defer IpPool.mux.Unlock()
	if mi, ok := macInfo[macAddr]; ok {
		if mi.Ip.Equal(ip) {
			mi.IsActive = false
			mi.LastLogin = time.Now()
			// 回写db数据
			dbdata.Set(dbdata.BucketMacIp, macAddr, mi)
		}
	}
}
