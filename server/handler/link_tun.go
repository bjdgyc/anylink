package handler

import (
	"fmt"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/coreos/go-iptables/iptables"
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
	cmdstr0 := fmt.Sprintln("modprobe -i tun")
	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %s multicast off", ifce.Name(), "1399")
	err = execCmd([]string{cmdstr0, cmdstr1})
	if err != nil {
		base.Fatal("testTun err: ", err)
	}
	// 开启服务器转发
	if err := execCmd([]string{"sysctl -w net.ipv4.ip_forward=1"}); err != nil {
		base.Fatal(err)
	}
	if base.Cfg.IptablesNat {
		// 添加NAT转发规则
		ipt, err := iptables.New()
		if err != nil {
			base.Fatal(err)
			return
		}

		// 修复 rockyos nat 不生效
		cmdstr0 := fmt.Sprintln("modprobe -i iptable_filter")
		cmdstr1 := fmt.Sprintf("modprobe -i iptable_nat")
		err = execCmd([]string{cmdstr0, cmdstr1})
		if err != nil {
			base.Fatal("testTun err: ", err)
		}

		natRule := []string{"-s", base.Cfg.Ipv4CIDR, "-o", base.Cfg.Ipv4Master, "-j", "MASQUERADE"}
		forwardRule := []string{"-j", "ACCEPT"}
		if natExists, _ := ipt.Exists("nat", "POSTROUTING", natRule...); !natExists {
			ipt.Insert("nat", "POSTROUTING", 1, natRule...)
		}
		if forwardExists, _ := ipt.Exists("filter", "FORWARD", forwardRule...); !forwardExists {
			ipt.Insert("filter", "FORWARD", 1, forwardRule...)
		}
		base.Info(ipt.List("nat", "POSTROUTING"))
		base.Info(ipt.List("filter", "FORWARD"))
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
	cSess.SetIfName(ifce.Name())

	cmdstr1 := fmt.Sprintf("ip link set dev %s up mtu %d multicast off alias %s.%s", ifce.Name(), cSess.Mtu,
		cSess.Group.Name, cSess.Username)
	cmdstr2 := fmt.Sprintf("ip addr add dev %s local %s peer %s/32",
		ifce.Name(), base.Cfg.Ipv4Gateway, cSess.IpAddr)
	err = execCmd([]string{cmdstr1, cmdstr2})
	if err != nil {
		base.Error(err)
		_ = ifce.Close()
		return err
	}

	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifce.Name())
	execCmd([]string{cmdstr3})

	go tunRead(ifce, cSess)
	go tunWrite(ifce, cSess)
	return nil
}

func tunWrite(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("LinkTun return", cSess.IpAddr)
		cSess.Close()
		_ = ifce.Close()
	}()

	var (
		err error
		pl  *sessdata.Payload
	)

	for {
		select {
		case pl = <-cSess.PayloadIn:
		case <-cSess.CloseChan:
			return
		}

		_, err = ifce.Write(pl.Data)
		if err != nil {
			base.Error("tun Write err", err)
			return
		}

		putPayloadInBefore(cSess, pl)
	}
}

func tunRead(ifce *water.Interface, cSess *sessdata.ConnSession) {
	defer func() {
		base.Debug("tunRead return", cSess.IpAddr)
		_ = ifce.Close()
	}()
	var (
		err error
		n   int
	)

	for {
		// data := make([]byte, BufferSize)
		pl := getPayload()
		n, err = ifce.Read(pl.Data)
		if err != nil {
			base.Error("tun Read err", n, err)
			return
		}

		// 更新数据长度
		pl.Data = (pl.Data)[:n]

		// data = data[:n]
		// ip_src := waterutil.IPv4Source(data)
		// ip_dst := waterutil.IPv4Destination(data)
		// ip_port := waterutil.IPv4DestinationPort(data)
		// fmt.Println("sent:", ip_src, ip_dst, ip_port)
		// packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
		// fmt.Println("read:", packet)

		if payloadOut(cSess, pl) {
			return
		}
	}
}
