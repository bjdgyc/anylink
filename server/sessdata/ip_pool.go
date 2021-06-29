package sessdata

import (
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
)

var (
	IpPool   = &ipPoolConfig{}
	ipActive = map[string]bool{}
)

type ipPoolConfig struct {
	mux sync.Mutex
	// 计算动态ip
	Ipv4Gateway net.IP
	Ipv4Mask    net.IP
	Ipv4IPNet   *net.IPNet
	IpLongMin   uint32
	IpLongMax   uint32
}

func initIpPool() {

	// 地址处理
	_, ipNet, err := net.ParseCIDR(base.Cfg.Ipv4CIDR)
	if err != nil {
		panic(err)
	}
	IpPool.Ipv4IPNet = ipNet
	IpPool.Ipv4Mask = net.IP(ipNet.Mask)
	IpPool.Ipv4Gateway = net.ParseIP(base.Cfg.Ipv4Gateway)

	// 网络地址零值
	// zero := binary.BigEndian.Uint32(ip.Mask(mask))
	// 广播地址
	// one, _ := ipNet.Mask.Size()
	// max := min | uint32(math.Pow(2, float64(32-one))-1)

	// ip地址池
	IpPool.IpLongMin = ip2long(net.ParseIP(base.Cfg.Ipv4Start))
	IpPool.IpLongMax = ip2long(net.ParseIP(base.Cfg.Ipv4End))
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

// AcquireIp 获取动态ip
func AcquireIp(username, macAddr string) net.IP {
	IpPool.mux.Lock()
	defer IpPool.mux.Unlock()

	tNow := time.Now()

	// 判断已经分配过
	mi := &dbdata.IpMap{}
	err := dbdata.One("MacAddr", macAddr, mi)
	if err == nil {
		ip := mi.IpAddr
		ipStr := ip.String()
		// 跳过活跃连接
		_, ok := ipActive[ipStr]
		// 检测原有ip是否在新的ip池内
		if IpPool.Ipv4IPNet.Contains(ip) && !ok {
			mi.Username = username
			mi.LastLogin = tNow
			// 回写db数据
			_ = dbdata.Save(mi)
			ipActive[ipStr] = true
			return ip
		}

		_ = dbdata.Del(mi)

	}

	farIp := &dbdata.IpMap{LastLogin: tNow}
	// 全局遍历超过租期ip
	for i := IpPool.IpLongMin; i <= IpPool.IpLongMax; i++ {
		ip := long2ip(i)
		ipStr := ip.String()

		// 跳过活跃连接
		if _, ok := ipActive[ipStr]; ok {
			continue
		}

		v := &dbdata.IpMap{}
		err = dbdata.One("IpAddr", ip, v)
		if err != nil {
			if dbdata.CheckErrNotFound(err) {
				// 该ip没有被使用
				mi = &dbdata.IpMap{IpAddr: ip, MacAddr: macAddr, Username: username, LastLogin: tNow}
				_ = dbdata.Save(mi)
				ipActive[ipStr] = true
				return ip
			}
			base.Error(err)
			return nil
		}

		// 跳过ip保留
		if v.Keep {
			continue
		}

		// 已经超过租期
		if tNow.Sub(v.LastLogin) > time.Duration(base.Cfg.IpLease)*time.Second {
			_ = dbdata.Del(v)
			mi = &dbdata.IpMap{IpAddr: ip, MacAddr: macAddr, Username: username, LastLogin: tNow}
			// 重写db数据
			_ = dbdata.Save(mi)
			ipActive[ipStr] = true
			return ip
		}

		// 其他情况判断最早登陆
		if v.LastLogin.Before(farIp.LastLogin) {
			farIp = v
		}
	}

	// 全都在线，没有数据可用
	if farIp.Id == 0 {
		return nil
	}

	// 使用最早登陆的mac ip
	ip := farIp.IpAddr
	ipStr := ip.String()
	mi = &dbdata.IpMap{IpAddr: ip, MacAddr: macAddr, Username: username, LastLogin: tNow}
	// 回写db数据
	_ = dbdata.Save(mi)
	ipActive[ipStr] = true
	return ip
}

// 回收ip
func ReleaseIp(ip net.IP, macAddr string) {
	IpPool.mux.Lock()
	defer IpPool.mux.Unlock()

	delete(ipActive, ip.String())
	mi := &dbdata.IpMap{}
	err := dbdata.One("IpAddr", ip, mi)
	if err == nil {
		mi.LastLogin = time.Now()
		_ = dbdata.Save(mi)
	}
}
