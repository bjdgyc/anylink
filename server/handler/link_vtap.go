package handler

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
)

// link vtap
const vTapPrefix = "lvtap"

type Vtap struct {
	*os.File
	ifName string
}

func (v *Vtap) Close() error {
	v.File.Close()
	cmdstr := fmt.Sprintf("ip link del %s", v.ifName)
	return execCmd([]string{cmdstr})
}

func checkMacvtap() {
	_setGateway()
	_checkTapIp(base.Cfg.Ipv4Master)

	ifName := "anylinkMacvtap"
	// 加载 macvtap
	cmdstr0 := fmt.Sprintf("modprobe -i macvtap")
	// 开启主网卡混杂模式
	cmdstr1 := fmt.Sprintf("ip link set dev %s promisc on", base.Cfg.Ipv4Master)
	// 测试 macvtap 功能
	cmdstr2 := fmt.Sprintf("ip link add link %s name %s type macvtap mode bridge", base.Cfg.Ipv4Master, ifName)
	cmdstr3 := fmt.Sprintf("ip link del %s", ifName)
	err := execCmd([]string{cmdstr0, cmdstr1, cmdstr2, cmdstr3})
	if err != nil {
		base.Fatal(err)
	}
}

// 创建 Macvtap 网卡
func LinkMacvtap(cSess *sessdata.ConnSession) error {
	capL := sessdata.IpPool.IpLongMax - sessdata.IpPool.IpLongMin
	ipN := utils.Ip2long(cSess.IpAddr) % capL
	ifName := fmt.Sprintf("%s%d", vTapPrefix, ipN)

	cSess.SetIfName(ifName)

	cmdstr1 := fmt.Sprintf("ip link add link %s name %s type macvtap mode bridge", base.Cfg.Ipv4Master, ifName)
	cmdstr2 := fmt.Sprintf("ip link set dev %s up mtu %d address %s", ifName, cSess.Mtu, cSess.MacHw)
	err := execCmd([]string{cmdstr1, cmdstr2})
	if err != nil {
		base.Error(err)
		return err
	}
	cmdstr3 := fmt.Sprintf("sysctl -w net.ipv6.conf.%s.disable_ipv6=1", ifName)
	execCmd([]string{cmdstr3})

	return createVtap(cSess, ifName)
}

func checkIpvtap() {

}

// 创建 Ipvtap 网卡
func LinkIpvtap(cSess *sessdata.ConnSession) error {
	return nil
}

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func createVtap(cSess *sessdata.ConnSession, ifName string) error {
	// 初始化 ifName
	inf, err := net.InterfaceByName(ifName)
	if err != nil {
		base.Error(err)
		return err
	}

	tName := fmt.Sprintf("/dev/tap%d", inf.Index)

	var fdInt int

	fdInt, err = syscall.Open(tName, os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}

	var flags uint16 = syscall.IFF_TAP | syscall.IFF_NO_PI
	var req ifReq
	req.Flags = flags

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fdInt),
		uintptr(syscall.TUNSETIFF),
		uintptr(unsafe.Pointer(&req)),
	)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}

	file := os.NewFile(uintptr(fdInt), tName)
	ifce := &Vtap{file, ifName}

	go allTapRead(ifce, cSess)
	go allTapWrite(ifce, cSess)
	return nil
}

// 销毁未关闭的vtap
func destroyVtap() {
	its, err := net.Interfaces()
	if err != nil {
		base.Error(err)
		return
	}
	for _, v := range its {
		if strings.HasPrefix(v.Name, vTapPrefix) {
			// 删除原来的网卡
			cmdstr := fmt.Sprintf("ip link del %s", v.Name)
			execCmd([]string{cmdstr})
		}
	}
}
