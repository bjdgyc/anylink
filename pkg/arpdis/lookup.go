// Currently only Darwin and Linux support this.

package arpdis

import (
	"log"
	"net"
	"os/exec"
	"strings"
)

func doLookup(ip net.IP) *Addr {
	// ping := exec.Command("ping", "-c1", "-t1", ip.String())
	// if err := ping.Run(); err != nil {
	// 	addr := &Addr{IP: ip, Type: TypeUnreachable}
	// 	return addr
	// }

	err := doPing(ip.String())
	if err != nil {
		// log.Println(err)
		addr := &Addr{IP: ip, Type: TypeUnreachable}
		return addr
	}

	return doArpShow(ip)
}

func doArpShow(ip net.IP) *Addr {
	cmd := exec.Command("ip", "n", "show", ip.String())
	out, err := cmd.Output()
	if err != nil {
		log.Println("lookup show", err)
		return nil
	}

	// os.Open("/proc/net/arp")
	// 192.168.1.2      0x1         0x2         e0:94:67:e2:42:5d     *        eth0
	// 192.168.1.2 dev eth0 lladdr 08:00:27:94:a5:a4 STALE
	outS := strings.ReplaceAll(string(out), "  ", " ")
	outS = strings.TrimSpace(outS)
	arpArr := strings.Split(outS, " ")
	if len(arpArr) != 6 {
		log.Println("lookup arpArr", outS, ip)
		return nil
	}
	mac, err := net.ParseMAC(arpArr[4])
	if err != nil {
		log.Println("lookup mac", outS, err)
		return nil
	}

	return &Addr{IP: ip, HardwareAddr: mac}
}

// IP address       HW type     Flags       HW address            Mask     Device
// 172.23.24.12     0x1         0x2         00:e0:4c:73:5c:48     *        anylink0
// 172.23.24.1      0x1         0x2         3c:8c:40:a0:7a:2d     *        anylink0
// 172.23.24.13     0x1         0x2         00:1c:42:4d:33:46     *        anylink0
// 172.23.24.2      0x1         0x0         00:00:00:00:00:00     *        anylink0
// 172.23.24.14     0x1         0x0         00:00:00:00:00:00     *        anylink0
