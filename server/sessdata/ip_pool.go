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
	ipLease   = map[string]bool{}
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
	go cronIpLease()
}

func cronIpLease() {
	getIpLease()
	tick := time.NewTicker(time.Minute * 30)
	for range tick.C {
		getIpLease()
	}
}

func getIpLease() {
	xdb := dbdata.GetXdb()
	keepIpMaps := []dbdata.IpMap{}
	sNow := time.Now().Add(-1 * time.Duration(base.Cfg.IpLease) * time.Second)
	err := xdb.Cols("ip_addr").Where("keep=?", true).Or("last_login>?", sNow).Find(&keepIpMaps)
	if err != nil {
		base.Error(err)
	}
	// fmt.Println(keepIpMaps)
	ipPoolMux.Lock()
	ipLease = map[string]bool{}
	for _, v := range keepIpMaps {
		ipLease[v.IpAddr] = true
	}
	ipPoolMux.Unlock()
}

// AcquireIp 获取动态ip
func AcquireIp(username, macAddr string) net.IP {
	ipPoolMux.Lock()
	defer ipPoolMux.Unlock()

	tNow := time.Now()

	// 判断是否已经分配过
	mi := &dbdata.IpMap{}
	err := dbdata.One("mac_addr", macAddr, mi)
	// 存在ip记录
	if err == nil {
		ipStr := mi.IpAddr
		ip := net.ParseIP(ipStr)
		// 跳过活跃连接
		_, ok := ipActive[ipStr]
		// 检测原有ip是否在新的ip池内
		if IpPool.Ipv4IPNet.Contains(ip) && !ok &&
			utils.Ip2long(ip) >= IpPool.IpLongMin &&
			utils.Ip2long(ip) <= IpPool.IpLongMax {
			mi.Username = username
			mi.LastLogin = tNow
			// 回写db数据
			_ = dbdata.Set(mi)
			ipActive[ipStr] = true
			return ip
		}
		_ = dbdata.Del(mi)
	}

	// 全局遍历超过租期和未保留的ip
	for i := IpPool.IpLongMin; i <= IpPool.IpLongMax; i++ {
		ip := utils.Long2ip(i)
		ipStr := ip.String()

		// 跳过活跃连接
		if _, ok := ipActive[ipStr]; ok {
			continue
		}
		// 跳过ip租期内数据
		if _, ok := ipLease[ipStr]; ok {
			continue
		}

		v := &dbdata.IpMap{}
		err = dbdata.One("ip_addr", ipStr, v)
		if err == nil {
			// 存在记录直接跳过
			continue
		}

		if dbdata.CheckErrNotFound(err) {
			// 该ip没有被使用
			mi = &dbdata.IpMap{IpAddr: ipStr, MacAddr: macAddr, Username: username, LastLogin: tNow}
			_ = dbdata.Add(mi)
			ipActive[ipStr] = true
			return ip
		}
		// 查询报错
		base.Error(err)
		return nil
	}

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
	}
}
