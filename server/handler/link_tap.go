package handler

import (
	"fmt"
	"io"
	"net"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/arpdis"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

const bridgeName = "anylink0"

var (
	// 网关mac地址
	gatewayHw net.HardwareAddr
)

type LinkDriver interface {
	io.ReadWriteCloser
	Name() string
}

func _setGateway() {
	dstAddr := arpdis.Lookup(sessdata.IpPool.Ipv4Gateway, false)
	gatewayHw = dstAddr.HardwareAddr
	// 设置为静态地址映射
	dstAddr.Type = arpdis.TypeStatic
	arpdis.Add(dstAddr)
}

func _checkTapIp(ifName string) {
	iFace, err := net.InterfaceByName(ifName)
	if err != nil {
		base.Fatal("testTap err: ", err)
	}

	var ifIp net.IP

	addrs, err := iFace.Addrs()
	if err != nil {
		base.Fatal("testTap err: ", err)
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip.To4() == nil {
			continue
		}
		ifIp = ip
	}

	if !sessdata.IpPool.Ipv4IPNet.Contains(ifIp) {
		base.Fatal("tapIp or Ip network err")
	}
}

func checkTap() {
	_setGateway()
	_checkTapIp(bridgeName)
}

// 创建tap网卡
func LinkTap(cSess *sessdata.ConnSession) error {
	cfg := water.Config{
		DeviceType: water.TAP,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		base.Error(err)
		return err
	}

	cSess.SetIfName(ifce.Name())

	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast on", ifce.Name(), cSess.Mtu)
	cmdstr2 := fmt.Sprintf("ip link set dev %s master %s", ifce.Name(), bridgeName)
	err = execCmd([]string{cmdstr1, cmdstr2})
	if err != nil {
		base.Error(err)
		_ = ifce.Close()
		return err
	}

	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	execCmd([]string{cmdstr3})

	go allTapRead(ifce, cSess)
	go allTapWrite(ifce, cSess)
	return nil
}

// ========================通用代码===========================

func allTapWrite(ifce LinkDriver, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("LinkTap return", cSess.IpAddr)
		cSess.Close()
		ifce.Close()
	}()

	var (
		err   error
		dstHw net.HardwareAddr
		pl    *sessdata.Payload
		frame = make(ethernet.Frame, BufferSize)
		ipDst = net.IPv4(1, 2, 3, 4)
	)

	for {
		frame.Resize(BufferSize)

		select {
		case pl = <-cSess.PayloadIn:
		case <-cSess.CloseChan:
			return
		}

		// var frame ethernet.Frame
		switch pl.LType {
		default:
			// log.Println(payload)
		case sessdata.LTypeEthernet:
			copy(frame, pl.Data)
			frame = frame[:len(pl.Data)]

			// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
			// fmt.Println("wirteArp:", packet)
		case sessdata.LTypeIPData: // 需要转换成 Ethernet 数据
			ipSrc := waterutil.IPv4Source(pl.Data)
			if !ipSrc.Equal(cSess.IpAddr) {
				// 非分配给客户端ip，直接丢弃
				continue
			}

			if waterutil.IsIPv6(pl.Data) {
				// 过滤掉IPv6的数据
				continue
			}

			// packet := gopacket.NewPacket(pl.Data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("get:", packet)

			// 手动设置ipv4地址
			ipDst[12] = pl.Data[16]
			ipDst[13] = pl.Data[17]
			ipDst[14] = pl.Data[18]
			ipDst[15] = pl.Data[19]

			dstHw = gatewayHw
			if sessdata.IpPool.Ipv4IPNet.Contains(ipDst) {
				dstAddr := arpdis.Lookup(ipDst, true)
				// fmt.Println("dstAddr", dstAddr)
				if dstAddr != nil {
					dstHw = dstAddr.HardwareAddr
				}
			}

			// fmt.Println("Gateway", ipSrc, ipDst, dstHw)
			frame.Prepare(dstHw, cSess.MacHw, ethernet.NotTagged, ethernet.IPv4, len(pl.Data))
			copy(frame[12+2:], pl.Data)
		}

		// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
		// fmt.Println("write:", packet)
		_, err = ifce.Write(frame)
		if err != nil {
			base.Error("tap Write err", err)
			return
		}

		putPayload(pl)
	}
}

func allTapRead(ifce LinkDriver, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("tapRead return", cSess.IpAddr)
		ifce.Close()
	}()

	var (
		err   error
		n     int
		data  []byte
		frame = make(ethernet.Frame, BufferSize)
	)

	for {
		frame.Resize(BufferSize)

		n, err = ifce.Read(frame)
		if err != nil {
			base.Error("tap Read err", n, err)
			return
		}
		frame = frame[:n]

		switch frame.Ethertype() {
		default:
			continue
		case ethernet.IPv6:
			continue
		case ethernet.IPv4:
			// 发送IP数据
			data = frame.Payload()

			ip_dst := waterutil.IPv4Destination(data)
			if !ip_dst.Equal(cSess.IpAddr) {
				// 过滤非本机地址
				// log.Println(ip_dst, sess.Ip)
				continue
			}

			// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("put:", packet)

			pl := getPayload()
			// 拷贝数据到pl
			copy(pl.Data, data)
			// 更新切片长度
			pl.Data = pl.Data[:len(data)]
			if payloadOut(cSess, pl) {
				return
			}

		case ethernet.ARP:
			// 暂时仅实现了ARP协议
			packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.NoCopy)
			layer := packet.Layer(layers.LayerTypeARP)
			arpReq := layer.(*layers.ARP)

			if !cSess.IpAddr.Equal(arpReq.DstProtAddress) {
				// 过滤非本机地址
				continue
			}

			// fmt.Println("arp", time.Now(), net.IP(arpReq.SourceProtAddress), cSess.IpAddr)
			// fmt.Println(packet)

			// 返回ARP数据
			src := &arpdis.Addr{IP: cSess.IpAddr, HardwareAddr: cSess.MacHw}
			dst := &arpdis.Addr{IP: arpReq.SourceProtAddress, HardwareAddr: arpReq.SourceHwAddress}
			data, err = arpdis.NewARPReply(src, dst)
			if err != nil {
				base.Error(err)
				return
			}

			// 从接受的arp信息添加arp地址
			addr := &arpdis.Addr{
				IP:           append([]byte{}, dst.IP...),
				HardwareAddr: append([]byte{}, dst.HardwareAddr...),
			}
			arpdis.Add(addr)

			pl := getPayload()
			// 设置为二层数据类型
			pl.LType = sessdata.LTypeEthernet
			// 拷贝数据到pl
			copy(pl.Data, data)
			// 更新切片长度
			pl.Data = pl.Data[:len(data)]

			if payloadIn(cSess, pl) {
				return
			}

		}
	}
}
