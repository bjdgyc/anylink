package sessdata

import (
	"bytes"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/pkg/utils"
)

type Online struct {
	Token             string    `json:"token"`
	Username          string    `json:"username"`
	Group             string    `json:"group"`
	MacAddr           string    `json:"mac_addr"`
	UniqueMac         bool      `json:"unique_mac"`
	Ip                net.IP    `json:"ip"`
	RemoteAddr        string    `json:"remote_addr"`
	TransportProtocol string    `json:"transport_protocol"`
	TunName           string    `json:"tun_name"`
	Mtu               int       `json:"mtu"`
	Client            string    `json:"client"`
	BandwidthUp       string    `json:"bandwidth_up"`
	BandwidthDown     string    `json:"bandwidth_down"`
	BandwidthUpAll    string    `json:"bandwidth_up_all"`
	BandwidthDownAll  string    `json:"bandwidth_down_all"`
	LastLogin         time.Time `json:"last_login"`
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
	return GetOnlineSess("", "", false)
}

/**
 * @Description: GetOnlineSess
 * @param search_cate 分类：用户名、登录组、MAC地址、IP地址、远端地址
 * @param search_text 关键字，模糊搜索
 * @param show_sleeper 是否显示休眠用户
 * @return []Online
 */
func GetOnlineSess(search_cate string, search_text string, show_sleeper bool) []Online {
	var datas Onlines
	if strings.TrimSpace(search_text) == "" {
		search_cate = ""
	}
	sessMux.Lock()
	defer sessMux.Unlock()
	for _, v := range sessions {
		v.mux.Lock()
		cSess := v.CSess
		if cSess == nil {
			cSess = &ConnSession{}
		}
		// 选择需要比较的字符串
		var compareText string
		switch search_cate {
		case "username":
			compareText = v.Username
		case "group":
			compareText = v.Group
		case "mac_addr":
			compareText = v.MacAddr
		case "ip":
			if cSess != nil {
				compareText = cSess.IpAddr.String()
			}
		case "remote_addr":
			if cSess != nil {
				compareText = cSess.RemoteAddr
			}
		}
		if search_cate != "" && !strings.Contains(compareText, search_text) {
			v.mux.Unlock()
			continue
		}

		if show_sleeper || v.IsActive {
			transportProtocol := "TCP"
			dSess := cSess.GetDtlsSession()
			if dSess != nil {
				transportProtocol = "UDP"
			}
			val := Online{
				Token:             v.Token,
				Ip:                cSess.IpAddr,
				Username:          v.Username,
				Group:             v.Group,
				MacAddr:           v.MacAddr,
				UniqueMac:         v.UniqueMac,
				RemoteAddr:        cSess.RemoteAddr,
				TransportProtocol: transportProtocol,
				TunName:           cSess.IfName,
				Mtu:               cSess.Mtu,
				Client:            cSess.Client,
				BandwidthUp:       utils.HumanByte(cSess.BandwidthUpPeriod.Load()) + "/s",
				BandwidthDown:     utils.HumanByte(cSess.BandwidthDownPeriod.Load()) + "/s",
				BandwidthUpAll:    utils.HumanByte(cSess.BandwidthUpAll.Load()),
				BandwidthDownAll:  utils.HumanByte(cSess.BandwidthDownAll.Load()),
				LastLogin:         v.LastLogin,
			}
			datas = append(datas, val)
		}
		v.mux.Unlock()
	}
	sort.Sort(&datas)
	return datas
}
