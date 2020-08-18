package handler

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/bjdgyc/anylink/common"
	"github.com/songgao/water"
)

func testTun() {
	// 测试tun
	cfg := water.Config{
		DeviceType: water.TUN,
	}

	ifce, err := water.New(cfg)
	if err != nil {
		log.Fatal("open tun err: ", err)
	}
	// 测试ip命令
	cmdstr := fmt.Sprintf("ip link set dev %s up mtu %s multicast off", ifce.Name(), "1399")
	err = execCmd([]string{cmdstr})
	if err != nil {
		log.Fatal("ip cmd err: ", err)
	}
	ifce.Close()
}

// 创建tun网卡
func LinkTun(sess *Session) {
	defer func() {
		sess.Close()
		log.Println("LinkTun return")
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

	// arp on
	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %s multicast off", ifce.Name(), sess.Mtu)
	cmdstr2 := fmt.Sprintf("ip addr add dev %s local %s peer %s/32",
		ifce.Name(), common.ServerCfg.Ipv4GateWay, sess.NetIp)
	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	cmdStrs := []string{cmdstr1, cmdstr2, cmdstr3}
	err = execCmd(cmdStrs)
	if err != nil {
		return
	}

	go tunRead(ifce, sess)

	var payload *Payload

	for {
		select {
		case payload = <-sess.PayloadIn:
		case <-sess.Closed:
			return
		}

		// ip_src := waterutil.IPv4Source(payload.data)
		// ip_des := waterutil.IPv4Destination(payload.data)
		// ip_port := waterutil.IPv4DestinationPort(payload.data)
		// fmt.Println("write: ", ip_src, ip_des.String(), ip_port, len(payload.data))

		_, err = ifce.Write(payload.data)
		if err != nil {
			log.Println("tun Write err", err)
			return
		}
	}

}

func tunRead(ifce *water.Interface, sess *Session) {
	var (
		err error
		n   int
	)

	for {
		packet := make([]byte, 1500)
		n, err = ifce.Read(packet)
		if err != nil {
			log.Println("tun Read err", n, err)
			return
		}

		payload := &Payload{
			ptype: 0x00,
			data:  packet[:n],
		}

		select {
		case sess.PayloadOut <- payload:
		case <-sess.Closed:
			return
		}
	}
}

func execCmd(cmdStrs []string) error {
	for _, cmdStr := range cmdStrs {
		cmd := exec.Command("bash", "-c", cmdStr)
		b, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(b), err)
			return err
		}
	}
	return nil
}
