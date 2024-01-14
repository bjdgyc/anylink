package sessdata

import (
	"net"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
)

var (
	IpPool   = &ipPoolConfig{}
	ipActive = map[string]bool{}
	// ipKeep and ipLease  ipAddr => type
	// ipLease   = map[string]bool{}
	ipPoolMux sync.Mutex
	// 记录循环点
	loopCurIp uint32
)

type ipPoolConfig struct {
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

	ipv4Gateway := net.ParseIP(base.Cfg.Ipv4Gateway)
	ipStart := net.ParseIP(base.Cfg.Ipv4Start)
	ipEnd := net.ParseIP(base.Cfg.Ipv4End)
	if !ipNet.Contains(ipv4Gateway) || !ipNet.Contains(ipStart) || !ipNet.Contains(ipEnd) {
		panic("ip段 设置错误")
	}
	// ip地址池
	IpPool.Ipv4Gateway = ipv4Gateway
	IpPool.IpLongMin = utils.Ip2long(ipStart)
	IpPool.IpLongMax = utils.Ip2long(ipEnd)

	loopCurIp = IpPool.IpLongMin

	// 网络地址零值
	// zero := binary.BigEndian.Uint32(ip.Mask(mask))
	// 广播地址
	// one, _ := ipNet.Mask.Size()
	// max := min | uint32(math.Pow(2, float64(32-one))-1)

	// 获取IpLease数据
	// go cronIpLease()
}

// func cronIpLease() {
// 	getIpLease()
// 	tick := time.NewTicker(time.Minute * 30)
// 	for range tick.C {
// 		getIpLease()
// 	}
// }
//
// func getIpLease() {
// 	xdb := dbdata.GetXdb()
// 	keepIpMaps := []dbdata.IpMap{}
// 	sNow := time.Now().Add(-1 * time.Duration(base.Cfg.IpLease) * time.Second)
// 	err := xdb.Cols("ip_addr").Where("keep=?", true).
// 		Or("unique_mac=? and last_login>?", true, sNow).Find(&keepIpMaps)
// 	if err != nil {
// 		base.Error(err)
// 	}
// 	// fmt.Println(keepIpMaps)
// 	ipPoolMux.Lock()
// 	ipLease = map[string]bool{}
// 	for _, v := range keepIpMaps {
// 		ipLease[v.IpAddr] = true
// 	}
// 	ipPoolMux.Unlock()
// }

// AcquireIp 获取动态ip
func AcquireIp(username, macAddr string, uniqueMac bool) (newIp net.IP) {
	base.Trace("AcquireIp start:", username, macAddr, uniqueMac)
	ipPoolMux.Lock()
	defer func() {
		ipPoolMux.Unlock()
		base.Trace("AcquireIp end:", username, macAddr, uniqueMac, newIp)
	}()

	var (
		err  error
		tNow = time.Now()
	)

	if uniqueMac {
		// 判断是否已经分配过
		mi := &dbdata.IpMap{}
		err = dbdata.One("mac_addr", macAddr, mi)
		if err != nil {
			// 没有查询到数据
			if dbdata.CheckErrNotFound(err) {
				return loopIp(username, macAddr, uniqueMac)
			}
			// 查询报错
			base.Error(err)
			return nil
		}

		// 存在ip记录
		base.Trace("uniqueMac:", username, mi)
		ipStr := mi.IpAddr
		ip := net.ParseIP(ipStr)
		// 跳过活跃连接
		_, ok := ipActive[ipStr]
		// 检测原有ip是否在新的ip池内
		// IpPool.Ipv4IPNet.Contains(ip) &&
		if !ok &&
			utils.Ip2long(ip) >= IpPool.IpLongMin &&
			utils.Ip2long(ip) <= IpPool.IpLongMax {
			mi.Username = username
			mi.LastLogin = tNow
			mi.UniqueMac = uniqueMac
			// 回写db数据
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			return ip
		}
		// 删除当前macAddr
		mi = &dbdata.IpMap{MacAddr: macAddr}
		_ = dbdata.Del(mi)

	} else {
		// 没有获取到mac的情况
		ipMaps := []dbdata.IpMap{}
		err = dbdata.FindWhere(&ipMaps, 50, 1, "username=? and unique_mac=?", username, false)
		if err != nil {
			// 没有查询到数据
			if dbdata.CheckErrNotFound(err) {
				return loopIp(username, macAddr, uniqueMac)
			}
			// 查询报错
			base.Error(err)
			return nil
		}

		// 遍历mac记录
		for _, mi := range ipMaps {
			ipStr := mi.IpAddr
			ip := net.ParseIP(ipStr)

			// 跳过活跃连接
			if _, ok := ipActive[ipStr]; ok {
				continue
			}
			// 跳过保留ip
			if mi.Keep {
				continue
			}
			// 没有mac的 不需要验证租期
			// mi.LastLogin.Before(leaseTime) &&
			if utils.Ip2long(ip) >= IpPool.IpLongMin &&
				utils.Ip2long(ip) <= IpPool.IpLongMax {
				mi.LastLogin = tNow
				mi.MacAddr = macAddr
				mi.UniqueMac = uniqueMac
				// 回写db数据
				_ = dbdata.Set(mi)
				ipActive[ipStr] = true
				return ip
			}
		}
	}

	return loopIp(username, macAddr, uniqueMac)
}

func loopIp(username, macAddr string, uniqueMac bool) net.IP {
	var (
		i  uint32
		ip net.IP
	)

	i, ip = loopLong(loopCurIp, IpPool.IpLongMax, username, macAddr, uniqueMac)
	if ip != nil {
		loopCurIp = i
		return ip
	}

	i, ip = loopLong(IpPool.IpLongMin, loopCurIp, username, macAddr, uniqueMac)
	if ip != nil {
		loopCurIp = i
		return ip
	}

	base.Warn("no ip available, please see ip_map table row", username, macAddr)
	return nil
}

func loopLong(start, end uint32, username, macAddr string, uniqueMac bool) (uint32, net.IP) {
	var (
		err       error
		tNow      = time.Now()
		leaseTime = time.Now().Add(-1 * time.Duration(base.Cfg.IpLease) * time.Second)
	)

	// 全局遍历超过租期和未保留的ip
	for i := start; i <= end; i++ {
		ip := utils.Long2ip(i)
		ipStr := ip.String()

		// 跳过活跃连接
		if _, ok := ipActive[ipStr]; ok {
			continue
		}

		mi := &dbdata.IpMap{}
		err = dbdata.One("ip_addr", ipStr, mi)
		if err != nil {
			// 没有查询到数据
			if dbdata.CheckErrNotFound(err) {
				// 该ip没有被使用
				mi = &dbdata.IpMap{IpAddr: ipStr, MacAddr: macAddr, UniqueMac: uniqueMac, Username: username, LastLogin: tNow}
				_ = dbdata.Add(mi)
				ipActive[ipStr] = true
				return i, ip
			}
			// 查询报错
			base.Error(err)
			return 0, nil
		}

		// 查询到已经使用的ip
		// 跳过保留ip
		if mi.Keep {
			continue
		}
		// 判断租期
		if mi.LastLogin.Before(leaseTime) {
			// 存在记录，说明已经超过租期，可以直接使用
			mi.LastLogin = tNow
			mi.MacAddr = macAddr
			mi.UniqueMac = uniqueMac
			// 回写db数据
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			return i, ip
		}
	}

	return 0, nil
}

// 回收ip
func ReleaseIp(ip net.IP, macAddr string) {
	ipPoolMux.Lock()
	defer ipPoolMux.Unlock()

	delete(ipActive, ip.String())

	mi := &dbdata.IpMap{}
	err := dbdata.One("ip_addr", ip.String(), mi)
	if err == nil {
		mi.LastLogin = time.Now()
		_ = dbdata.Set(mi)
	}
}
