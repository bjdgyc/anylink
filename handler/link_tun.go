package handler

import (
	"fmt"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/songgao/water"
)

func checkTun() {
	// 测试tun
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		base.Fatal("open tun err: ", err)
	}
	defer ifce.Close()

	// 测试ip命令
	cmdstr := fmt.Sprintf("ip link set dev %s up mtu %s multicast off", ifce.Name(), "1399")
	err = execCmd([]string{cmdstr})
	if err != nil {
		base.Fatal("testTun err: ", err)
	}
}

// 创建tun网卡
func LinkTun(cSess *sessdata.ConnSession) error {
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		base.Error(err)
		return err
	}
	// log.Printf("Interface Name: %s\n", ifce.Name())
	cSess.SetTunName(ifce.Name())
	// cSess.TunName = ifce.Name()

	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast off", ifce.Name(), cSess.Mtu)
	cmdstr2 := fmt.Sprintf("ip addr add dev %s local %s peer %s/32",
		ifce.Name(), base.Cfg.Ipv4Gateway, cSess.IpAddr)
	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	cmdStrs := []string{cmdstr1, cmdstr2, cmdstr3}
	err = execCmd(cmdStrs)
	if err != nil {
		base.Error(err)
		ifce.Close()
		return err
	}

	go tunRead(ifce, cSess)
	go tunWrite(ifce, cSess)
	return nil
}

func tunWrite(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		// log.Println("LinkTun return")
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

		_, err = ifce.Write(payload.Data)
		if err != nil {
			base.Error("tun Write err", err)
			return
		}
	}
}

func tunRead(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		// log.Println("tunRead return")
		ifce.Close()
	}()
	var (
		err error
		n   int
	)

	for {
		data := make([]byte, BufferSize)
		n, err = ifce.Read(data)
		if err != nil {
			base.Error("tun Read err", n, err)
			return
		}

		data = data[:n]

		// ip_src := waterutil.IPv4Source(data)
		// ip_dst := waterutil.IPv4Destination(data)
		// ip_port := waterutil.IPv4DestinationPort(data)
		// fmt.Println("sent:", ip_src, ip_dst, ip_port)
		// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
		// fmt.Println("read:", packet)

		if payloadOut(cSess, sessdata.LTypeIPData, 0x00, data) {
			return
		}

	}
}
