package handler

import (
	"fmt"
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
	bridgeIp net.IP
	bridgeHw net.HardwareAddr
)

func checkTap() {
	brFace, err := net.InterfaceByName(bridgeName)
	if err != nil {
		base.Fatal("testTap err: ", err)
	}
	bridgeHw = brFace.HardwareAddr

	addrs, err := brFace.Addrs()
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip.To4() == nil {
			continue
		}
		bridgeIp = ip
	}
	if bridgeIp == nil && bridgeHw == nil {
		base.Fatal("bridgeIp is err")
	}

	if !sessdata.IpPool.Ipv4IPNet.Contains(bridgeIp) {
		base.Fatal("bridgeIp or Ip network err")
	}
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

	cSess.TunName = ifce.Name()

	// arp on
	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast on", ifce.Name(), cSess.Mtu)
	cmdstr2 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	cmdstr3 := fmt.Sprintf("ip link set dev %s master %s", ifce.Name(), bridgeName)
	cmdStrs := []string{cmdstr1, cmdstr2, cmdstr3}
	err = execCmd(cmdStrs)
	if err != nil {
		base.Error(err)
		ifce.Close()
		return err
	}

	go tapRead(ifce, cSess)
	go tapWrite(ifce, cSess)
	return nil
}

func tapWrite(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		// log.Println("LinkTap return")
		cSess.Close()
		ifce.Close()
	}()

	var (
		err     error
		payload *sessdata.Payload
	)

	for {
		select {
		case payload = <-cSess.PayloadIn:
		case <-cSess.CloseChan:
			return
		}

		var frame ethernet.Frame
		switch payload.LType {
		default:
			// log.Println(payload)
		case sessdata.LTypeEthernet:
			frame = payload.Data
		case sessdata.LTypeIPData: // 需要转换成 Ethernet 数据
			data := payload.Data

			ip_src := waterutil.IPv4Source(data)
			if waterutil.IsIPv6(data) || !ip_src.Equal(cSess.IpAddr) {
				// 过滤掉IPv6的数据
				// 非分配给客户端ip，直接丢弃
				continue
			}

			// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("get:", packet)

			ip_dst := waterutil.IPv4Destination(data)
			// fmt.Println("get:", ip_src, ip_dst)

			var dstHw net.HardwareAddr
			if !sessdata.IpPool.Ipv4IPNet.Contains(ip_dst) || ip_dst.Equal(sessdata.IpPool.Ipv4Gateway) {
				// 不是同一网段，使用网关mac地址
				dstAddr := arpdis.Lookup(sessdata.IpPool.Ipv4Gateway, false)
				dstHw = dstAddr.HardwareAddr
			} else {
				dstAddr := arpdis.Lookup(ip_dst, true)
				// fmt.Println("dstAddr", dstAddr)
				if dstAddr != nil {
					dstHw = dstAddr.HardwareAddr
				} else {
					dstHw = bridgeHw
				}

			}
			// fmt.Println("Gateway", ip_dst, dstAddr.HardwareAddr)

			frame.Prepare(dstHw, cSess.MacHw, ethernet.NotTagged, ethernet.IPv4, len(data))
			copy(frame[12+2:], data)
		}

		// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
		// fmt.Println("write:", packet)
		_, err = ifce.Write(frame)
		if err != nil {
			base.Error("tap Write err", err)
			return
		}
	}
}

func tapRead(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		// log.Println("tapRead return")
		ifce.Close()
	}()

	var (
		err error
		n   int
		buf []byte
	)

	for {
		var frame ethernet.Frame
		frame.Resize(BufferSize)
		n, err = ifce.Read(frame)
		if err != nil {
			base.Error("tap Read err", n, err)
			return
		}
		frame = frame[:n]

		switch frame.Ethertype() {
		default:
			// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
			// fmt.Println(packet)
			continue
		case ethernet.IPv6:
			continue
		case ethernet.IPv4:
			// 发送IP数据
			data := frame.Payload()

			ip_dst := waterutil.IPv4Destination(data)
			if !ip_dst.Equal(cSess.IpAddr) {
				// 过滤非本机地址
				// log.Println(ip_dst, sess.Ip)
				continue
			}

			// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
			// fmt.Println("put:", packet)

			if payloadOut(cSess, sessdata.LTypeIPData, 0x00, data) {
				return
			}

		case ethernet.ARP:
			// 暂时仅实现了ARP协议
			packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
			layer := packet.Layer(layers.LayerTypeARP)
			arpReq := layer.(*layers.ARP)

			if !cSess.IpAddr.Equal(arpReq.DstProtAddress) {
				// 过滤非本机地址
				continue
			}

			// fmt.Println("arp", net.IP(arpReq.SourceProtAddress), sess.Ip)
			// fmt.Println(packet)

			// 返回ARP数据
			src := &arpdis.Addr{IP: cSess.IpAddr, HardwareAddr: cSess.MacHw}
			dst := &arpdis.Addr{IP: arpReq.SourceProtAddress, HardwareAddr: frame.Source()}
			buf, err = arpdis.NewARPReply(src, dst)
			if err != nil {
				base.Error(err)
				return
			}

			// 从接受的arp信息添加arp地址
			addr := &arpdis.Addr{
				IP:           make([]byte, len(arpReq.SourceProtAddress)),
				HardwareAddr: make([]byte, len(frame.Source())),
			}
			// addr.IP = arpReq.SourceProtAddress
			// addr.HardwareAddr = frame.Source()
			copy(addr.IP, arpReq.SourceProtAddress)
			copy(addr.HardwareAddr, frame.Source())
			arpdis.Add(addr)

			if payloadIn(cSess, sessdata.LTypeEthernet, 0x00, buf) {
				return
			}

		}
	}
}
