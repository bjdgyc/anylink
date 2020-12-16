package arpdis

import (
	"net"
	"sync"
	"time"
)

const (
	StaleTimeNormal      = time.Minute * 5
	StaleTimeUnreachable = time.Minute * 10

	TypeNormal      = 0
	TypeUnreachable = 1
	TypeStatic      = 2
)

var (
	table   = make(map[string]*Addr)
	tableMu sync.RWMutex
)

type Addr struct {
	IP           net.IP
	HardwareAddr net.HardwareAddr
	disTime      time.Time
	Type         int8
}

func Lookup(ip net.IP, onlyTable bool) *Addr {
	addr := tableLookup(ip.To4())
	if addr != nil || onlyTable {
		return addr
	}

	addr = doLookup(ip.To4())
	Add(addr)
	return addr
}

// Add adds a IP-MAC map to a runtime ARP table.
func tableLookup(ip net.IP) *Addr {
	tableMu.Lock()
	addr := table[ip.To4().String()]
	tableMu.Unlock()
	if addr == nil {
		return nil
	}

	// 判断老化过期时间
	tsub := time.Now().Sub(addr.disTime)
	switch addr.Type {
	case TypeNormal:
		if tsub > StaleTimeNormal {
			return nil
		}
	case TypeUnreachable:
		if tsub > StaleTimeUnreachable {
			return nil
		}
	case TypeStatic:
	}

	return addr
}

// Add adds a IP-MAC map to a runtime ARP table.
func Add(addr *Addr) {
	if addr == nil {
		return
	}
	if addr.disTime.IsZero() {
		addr.disTime = time.Now()
	}
	ip := addr.IP.To4().String()
	tableMu.Lock()
	defer tableMu.Unlock()
	if a, ok := table[ip]; ok {
		// 静态地址只能设置一次
		if a.Type == TypeStatic {
			return
		}
	}
	table[ip] = addr
}

// Delete removes an IP from the runtime ARP table.
func Delete(ip net.IP) {
	tableMu.Lock()
	defer tableMu.Unlock()
	delete(table, ip.To4().String())
}

// List returns the current runtime ARP table.
func List() map[string]*Addr {
	tableMu.RLock()
	defer tableMu.RUnlock()
	return table
}
