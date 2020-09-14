package handler

import (
	"fmt"
	"log"

	"github.com/bjdgyc/anylink/common"
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
		log.Fatal("open tun err: ", err)
	}
	defer ifce.Close()

	// 测试ip命令
	cmdstr := fmt.Sprintf("ip link set dev %s up mtu %s multicast off", ifce.Name(), "1399")
	err = execCmd([]string{cmdstr})
	if err != nil {
		log.Fatal("testTun err: ", err)
	}
}

// 创建tun网卡
func LinkTun(sess *sessdata.ConnSession) {
	defer func() {
		log.Println("LinkTun return")
		sess.Close()
	}()

	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	// log.Printf("Interface Name: %s\n", ifce.Name())
	sess.TunName = ifce.Name()
	defer ifce.Close()

	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast off", ifce.Name(), sess.Mtu)
	cmdstr2 := fmt.Sprintf("ip addr add dev %s local %s peer %s/32",
		ifce.Name(), common.ServerCfg.Ipv4Gateway, sess.Ip)
	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	cmdStrs := []string{cmdstr1, cmdstr2, cmdstr3}
	err = execCmd(cmdStrs)
	if err != nil {
		return
	}

	go tunRead(ifce, sess)

	var payload *sessdata.Payload

	for {
		select {
		case payload = <-sess.PayloadIn:
		case <-sess.CloseChan:
			return
		}

		_, err = ifce.Write(payload.Data)
		if err != nil {
			log.Println("tun Write err", err)
			return
		}
	}

}

func tunRead(ifce *water.Interface, sess *sessdata.ConnSession) {
	defer func() {
		log.Println("tunRead return")
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
			log.Println("tun Read err", n, err)
			return
		}

		data = data[:n]

		// ip_src := waterutil.IPv4Source(data)
		// ip_dst := waterutil.IPv4Destination(data)
		// ip_port := waterutil.IPv4DestinationPort(data)
		// fmt.Println("sent:", ip_src, ip_dst, ip_port)
		// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
		// fmt.Println("read:", packet)

		if payloadOut(sess, sessdata.LTypeIPData, 0x00, data) {
			return
		}

	}
}
