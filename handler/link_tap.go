package handler

import (
	"fmt"
	"log"
	"net"

	"github.com/bjdgyc/anylink/arpdis"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

const bridgeName = "anylink0"

func checkTap() {
	brFace, err := net.InterfaceByName(bridgeName)
	if err != nil {
		log.Fatal("testTap err: ", err)
	}
	bridgeHw := brFace.HardwareAddr
	var bridgeIp net.IP
	addrs, err := brFace.Addrs()
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil || ip.To4() == nil {
			continue
		}
		bridgeIp = ip
	}
	if bridgeIp == nil && bridgeHw == nil {
		log.Fatalln("bridgeIp is err")
	}

	if !sessdata.IpPool.Ipv4IPNet.Contains(bridgeIp) {
		log.Fatalln("bridgeIp or Ip network err")
	}

	// 设置本机ip arp为静态
	addr := &arpdis.Addr{IP: bridgeIp.To4(), HardwareAddr: bridgeHw, Type: arpdis.TypeStatic}
	arpdis.Add(addr)
}

// 创建tap网卡
func LinkTap(sess *sessdata.ConnSession) {
	defer func() {
		log.Println("LinkTap return")
		sess.Close()
	}()

	cfg := water.Config{
		DeviceType: water.TAP,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	sess.TunName = ifce.Name()
	defer ifce.Close()

	// arp on
	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast on", ifce.Name(), sess.Mtu)
	cmdstr2 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	cmdstr3 := fmt.Sprintf("ip link set dev %s master %s", ifce.Name(), bridgeName)
	cmdStrs := []string{cmdstr1, cmdstr2, cmdstr3}
	err = execCmd(cmdStrs)
	if err != nil {
		return
	}

	// TODO 测试
	// sess.MacHw, _ = net.ParseMAC("3c:8c:40:a0:6a:3d")

	go loopArp(sess)
	go tapRead(ifce, sess)

	var (
		payload *sessdata.Payload
	)

	for {
		select {
		case payload = <-sess.PayloadIn:
		case <-sess.CloseChan:
			return
		}

		var frame ethernet.Frame
		switch payload.LType {
		default:
			log.Println(payload)
		case sessdata.LTypeEthernet:
			frame = payload.Data
		case sessdata.LTypeIPData: // 需要转换成 Ethernet 数据
			data := payload.Data

			ip_src := waterutil.IPv4Source(data)
			if waterutil.IsIPv6(data) || !ip_src.Equal(sess.Ip) {
				// 过滤掉IPv6的数据
				// 非分配给客户端ip，直接丢弃
				continue
			}

			ip_dst := waterutil.IPv4Destination(data)
			// fmt.Println("get:", ip_src, ip_dst)

			var dstAddr *arpdis.Addr
			if !sessdata.IpPool.Ipv4IPNet.Contains(ip_dst) || ip_dst.Equal(sessdata.IpPool.Ipv4Gateway) {
				// 不是同一网段，使用网关mac地址
				ip_dst = sessdata.IpPool.Ipv4Gateway
				dstAddr = arpdis.Lookup(ip_dst, false)
				if dstAddr == nil {
					log.Println("Ipv4Gateway mac err", ip_dst)
					return
				}
				// fmt.Println("Gateway", ip_dst, dstAddr.HardwareAddr)
			} else {
				// 同一网段内的其他主机
				dstAddr = arpdis.Lookup(ip_dst, true)
				// fmt.Println("other", ip_src, ip_dst, dstAddr)
				if dstAddr == nil || dstAddr.Type == arpdis.TypeUnreachable {
					// 异步检测发送数据包
					select {
					case sess.PayloadArp <- payload:
					case <-sess.CloseChan:
						return
					default:
						// PayloadArp 容量已经满了
						log.Println("PayloadArp is full", sess.Ip, ip_dst)
					}
					continue
				}
			}

			frame.Prepare(dstAddr.HardwareAddr, sess.MacHw, ethernet.NotTagged, ethernet.IPv4, len(data))
			copy(frame[12+2:], data)
		}

		// packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.Default)
		// fmt.Println("write:", packet)
		_, err = ifce.Write(frame)
		if err != nil {
			log.Println("tap Write err", err)
			return
		}
	}

}

// 异步处理获取ip对应的mac地址的数据
func loopArp(sess *sessdata.ConnSession) {
	defer func() {
		log.Println("loopArp return")
	}()

	var (
		payload *sessdata.Payload
		dstAddr *arpdis.Addr
		ip_dst  net.IP
	)

	for {
		select {
		case payload = <-sess.PayloadArp:
		case <-sess.CloseChan:
			return
		}

		ip_dst = waterutil.IPv4Destination(payload.Data)
		dstAddr = arpdis.Lookup(ip_dst, false)
		// 不可达数据包
		if dstAddr == nil || dstAddr.Type == arpdis.TypeUnreachable {
			// 直接丢弃数据
			// fmt.Println("Lookup", ip_dst)
			continue
		}

		// 正常获取mac地址
		if payloadInData(sess, payload) {
			return
		}

	}
}

func tapRead(ifce *water.Interface, sess *sessdata.ConnSession) {
	defer func() {
		log.Println("tapRead return")
		ifce.Close()
	}()

	var (
		err error
		n   int
		buf []byte
	)
	fmt.Println(sess.MacHw)

	for {
		var frame ethernet.Frame
		frame.Resize(BufferSize)
		n, err = ifce.Read(frame)
		if err != nil {
			log.Println("tap Read err", n, err)
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
			if !ip_dst.Equal(sess.Ip) {
				// 过滤非本机地址
				// log.Println(ip_dst, sess.Ip)
				continue
			}

			if payloadOut(sess, sessdata.LTypeIPData, 0x00, data) {
				return
			}

		case ethernet.ARP:
			// 暂时仅实现了ARP协议
			packet := gopacket.NewPacket(frame, layers.LayerTypeEthernet, gopacket.NoCopy)
			layer := packet.Layer(layers.LayerTypeARP)
			arpReq := layer.(*layers.ARP)

			// fmt.Println("arp", net.IP(arpReq.SourceProtAddress), sess.Ip)
			if !sess.Ip.Equal(arpReq.DstProtAddress) {
				// 过滤非本机地址
				continue
			}

			// fmt.Println("arp", arpReq.SourceProtAddress, sess.Ip)
			// fmt.Println(packet)

			// 返回ARP数据
			src := &arpdis.Addr{IP: sess.Ip, HardwareAddr: sess.MacHw}
			dst := &arpdis.Addr{IP: arpReq.SourceProtAddress, HardwareAddr: frame.Source()}
			buf, err = arpdis.NewARPReply(src, dst)
			if err != nil {
				log.Println(err)
				return
			}

			// 从接受的arp信息添加arp地址
			addr := &arpdis.Addr{}
			copy(addr.IP, arpReq.SourceProtAddress)
			copy(addr.HardwareAddr, frame.Source())
			arpdis.Add(addr)

			if payloadIn(sess, sessdata.LTypeEthernet, 0x00, buf) {
				return
			}

		}
	}
}
