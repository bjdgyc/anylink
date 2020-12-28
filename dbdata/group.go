package dbdata

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/base"
)

const (
	Allow = "allow"
	Deny  = "deny"
)

type GroupLinkAcl struct {
	// 自上而下匹配 默认 allow * *
	Action string     `json:"action"` // allow、deny
	Val    string     `json:"val"`
	Port   uint8      `json:"port"`
	IpNet  *net.IPNet `json:"-"`
}

type ValData struct {
	Val string `json:"val"`
}

type Group struct {
	Id           int            `json:"id" storm:"id,increment"`
	Name         string         `json:"name" storm:"unique"`
	Note         string         `json:"note"`
	AllowLan     bool           `json:"allow_lan"`
	ClientDns    []ValData      `json:"client_dns"`
	RouteInclude []ValData      `json:"route_include"`
	RouteExclude []ValData      `json:"route_exclude"`
	LinkAcl      []GroupLinkAcl `json:"link_acl"`
	Bandwidth    int            `json:"bandwidth"` // 带宽限制
	Status       int8           `json:"status"`    // 1正常
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func GetGroupNames() []string {
	var datas []Group
	err := All(&datas, 0, 0)
	if err != nil {
		base.Error(err)
		return nil
	}
	var names []string
	for _, v := range datas {
		names = append(names, v.Name)
	}
	return names
}

func SetGroup(g *Group) error {
	var err error
	if g.Name == "" {
		return errors.New("用户组名错误")
	}

	// 判断数据
	clientDns := []ValData{}
	for _, v := range g.ClientDns {
		if v.Val != "" {
			clientDns = append(clientDns, v)
		}
	}
	if len(clientDns) == 0 {
		return errors.New("DNS错误")
	}
	g.ClientDns = clientDns

	routeInclude := []ValData{}
	for _, v := range g.RouteInclude {
		if v.Val != "" {
			v1, _ := parseIpNet(v.Val)
			vn := ValData{Val: v1}
			routeInclude = append(routeInclude, vn)
		}
	}
	g.RouteInclude = routeInclude
	routeExclude := []ValData{}
	for _, v := range g.RouteExclude {
		if v.Val != "" {
			v1, _ := parseIpNet(v.Val)
			vn := ValData{Val: v1}
			routeExclude = append(routeExclude, vn)
		}
	}
	g.RouteExclude = routeExclude
	// 转换数据
	linkAcl := []GroupLinkAcl{}
	for _, v := range g.LinkAcl {
		if v.Val != "" {
			v1, v2 := parseIpNet(v.Val)
			if v2 != nil {
				vn := v
				vn.Val = v1
				vn.IpNet = v2
				linkAcl = append(linkAcl, vn)
			}
		}
	}
	g.LinkAcl = linkAcl

	g.UpdatedAt = time.Now()
	err = Save(g)

	return err
}

func parseIpNet(s string) (string, *net.IPNet) {
	ips := strings.Split(s, "/")
	if len(ips) != 2 {
		return "", nil
	}
	ip := net.ParseIP(ips[0])
	mask := net.ParseIP(ips[1])

	if strings.Contains(ips[0], ".") {
		ip = ip.To4()
		mask = mask.To4()
	}

	ipmask := net.IPMask(mask)
	ip0 := ip.Mask(ipmask)

	ipNetS := fmt.Sprintf("%s/%s", ip0, mask)
	ipNet := &net.IPNet{IP: ip0, Mask: ipmask}

	return ipNetS, ipNet
}
