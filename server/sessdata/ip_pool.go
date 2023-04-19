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
	ipAll    = map[string]*dbdata.IpMap{}
	ipKeep   = map[string]bool{}
	// ipKeep and ipLease  ipAddr => type
	// ipLease   = map[string]bool{}
	ipPoolMux sync.Mutex
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
	IpPool.Ipv4Gateway = net.ParseIP(base.Cfg.Ipv4Gateway)

	// 网络地址零值
	// zero := binary.BigEndian.Uint32(ip.Mask(mask))
	// 广播地址
	// one, _ := ipNet.Mask.Size()
	// max := min | uint32(math.Pow(2, float64(32-one))-1)

	// ip地址池
	IpPool.IpLongMin = utils.Ip2long(net.ParseIP(base.Cfg.Ipv4Start))
	IpPool.IpLongMax = utils.Ip2long(net.ParseIP(base.Cfg.Ipv4End))

	// 获取IpLease数据
	// go cronIpLease()
	getIpMapAll()
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

func IpAllSet(v *dbdata.IpMap) {
	ipPoolMux.Lock()
	defer ipPoolMux.Unlock()

	ipAll[v.IpAddr] = v
	if v.Keep {
		ipKeep[v.IpAddr] = true
	} else {
		delete(ipKeep, v.IpAddr)
	}
}

func IpAllDel(v *dbdata.IpMap) {
	ipPoolMux.Lock()
	defer ipPoolMux.Unlock()

	delete(ipAll, v.IpAddr)
	delete(ipKeep, v.IpAddr)
}

func getIpMapAll() {
	xdb := dbdata.GetXdb()
	ipMaps := []dbdata.IpMap{}
	err := xdb.Find(&ipMaps)
	if err != nil {
		base.Fatal(err)
	}
	for _, v := range ipMaps {
		ipAll[v.IpAddr] = &v
		if v.Keep {
			ipKeep[v.IpAddr] = true
		}
	}
}

// AcquireIp 获取动态ip
func AcquireIp(username, macAddr string, uniqueMac bool) net.IP {
	ipPoolMux.Lock()
	defer ipPoolMux.Unlock()

	var (
		err  error
		tNow = time.Now()
	)

	// 唯一mac
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
		ipStr := mi.IpAddr
		ip := net.ParseIP(ipStr)
		// 跳过活跃连接
		_, ok := ipActive[ipStr]
		// 检测原有ip是否在新的ip池内
		// IpPool.Ipv4IPNet.Contains(ip) &&
		if utils.Ip2long(ip) >= IpPool.IpLongMin &&
			utils.Ip2long(ip) <= IpPool.IpLongMax && !ok {
			mi.Username = username
			mi.LastLogin = tNow
			mi.UniqueMac = uniqueMac
			// 回写db数据
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			ipAll[ipStr] = mi
			return ip
		}
		// 原有ip不符合
		_ = dbdata.Del(mi)
		delete(ipAll, ipStr)
		return loopIp(username, macAddr, uniqueMac)
	}

	// 没有获取到mac
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
	// 遍历 用户的 mac记录
	for _, im := range ipMaps {
		mi := &im
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
		// 没有租期限制
		if utils.Ip2long(ip) >= IpPool.IpLongMin &&
			utils.Ip2long(ip) <= IpPool.IpLongMax {
			mi.LastLogin = tNow
			mi.MacAddr = macAddr
			mi.UniqueMac = uniqueMac
			// 回写db数据
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			ipAll[ipStr] = mi
			return ip
		}
	}

	return loopIp(username, macAddr, uniqueMac)
}

// 遍历ip
func loopIp(username, macAddr string, uniqueMac bool) net.IP {
	var (
		tNow      = time.Now()
		leaseTime = time.Now().Add(-1 * time.Duration(base.Cfg.IpLease) * time.Second)
	)

	// 全局遍历超过租期和未保留的ip
	for i := IpPool.IpLongMin; i <= IpPool.IpLongMax; i++ {
		ip := utils.Long2ip(i)
		ipStr := ip.String()

		// 跳过活跃连接
		if _, ok := ipActive[ipStr]; ok {
			continue
		}
		// 跳过保留ip
		if _, ok := ipKeep[ipStr]; ok {
			continue
		}

		mi, ok := ipAll[ipStr]
		if !ok {
			// 该ip没有被使用 直接返回
			mi = &dbdata.IpMap{IpAddr: ipStr, MacAddr: macAddr, UniqueMac: uniqueMac, Username: username, LastLogin: tNow}
			_ = dbdata.Add(mi)
			ipActive[ipStr] = true
			ipAll[ipStr] = mi
			return ip
		}

		// 存在记录
		if mi.LastLogin.Before(leaseTime) {
			// 存在记录，说明已经超过租期，可以直接使用
			mi.LastLogin = tNow
			mi.MacAddr = macAddr
			mi.UniqueMac = uniqueMac
			// 回写db数据
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			ipAll[ipStr] = mi
			return ip
		}
	}

	// 最后从未超过租期的记录获取一个

	base.Warn("no ip available, please see ip_map table row")
	return nil
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
		ipAll[mi.IpAddr] = mi
	}
}
