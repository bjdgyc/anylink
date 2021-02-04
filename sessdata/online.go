package sessdata

import (
	"bytes"
	"net"
	"sort"
	"sync/atomic"
	"time"

	"github.com/bjdgyc/anylink/pkg/utils"
)

type Online struct {
	Token            string    `json:"token"`
	Username         string    `json:"username"`
	Group            string    `json:"group"`
	MacAddr          string    `json:"mac_addr"`
	Ip               net.IP    `json:"ip"`
	RemoteAddr       string    `json:"remote_addr"`
	TunName          string    `json:"tun_name"`
	Mtu              int       `json:"mtu"`
	Client           string    `json:"client"`
	BandwidthUp      string    `json:"bandwidth_up"`
	BandwidthDown    string    `json:"bandwidth_down"`
	BandwidthUpAll   string    `json:"bandwidth_up_all"`
	BandwidthDownAll string    `json:"bandwidth_down_all"`
	LastLogin        time.Time `json:"last_login"`
}

type Onlines []Online

func (o Onlines) Len() int {
	return len(o)
}

func (o Onlines) Less(i, j int) bool {
	return bytes.Compare(o[i].Ip, o[j].Ip) < 0
}

func (o Onlines) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func OnlineSess() []Online {
	var datas Onlines
	sessMux.Lock()
	for _, v := range sessions {
		v.mux.Lock()
		if v.IsActive {
			val := Online{
				Token:            v.Token,
				Ip:               v.CSess.IpAddr,
				Username:         v.Username,
				Group:            v.Group,
				MacAddr:          v.MacAddr,
				RemoteAddr:       v.CSess.RemoteAddr,
				TunName:          v.CSess.TunName,
				Mtu:              v.CSess.Mtu,
				Client:           v.CSess.Client,
				BandwidthUp:      utils.HumanByte(atomic.LoadUint32(&v.CSess.BandwidthUpPeriod)) + "/s",
				BandwidthDown:    utils.HumanByte(atomic.LoadUint32(&v.CSess.BandwidthDownPeriod)) + "/s",
				BandwidthUpAll:   utils.HumanByte(atomic.LoadUint32(&v.CSess.BandwidthUpAll)),
				BandwidthDownAll: utils.HumanByte(atomic.LoadUint32(&v.CSess.BandwidthDownAll)),
				LastLogin:        v.LastLogin,
			}
			datas = append(datas, val)
		}
		v.mux.Unlock()
	}
	sessMux.Unlock()
	sort.Sort(&datas)
	return datas
}
