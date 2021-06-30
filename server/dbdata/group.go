package dbdata

import (
	"errors"
	"fmt"
	"net"
	"time"
)

func GetGroupNames() []string {
	var strings []string
	serr := x.Table("group").Cols("name").Find(&strings)
	if serr != nil {
		panic(serr)
	}
	//fmt.Println(strings)

	return strings
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
		return errors.New("DNS 错误")
	}
	g.ClientDns = clientDns

	routeInclude := []ValData{}
	for _, v := range g.RouteInclude {
		if v.Val != "" {
			ipMask, _, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("RouteInclude 错误" + err.Error())
			}

			v.IpMask = ipMask
			routeInclude = append(routeInclude, v)
		}
	}
	g.RouteInclude = routeInclude
	routeExclude := []ValData{}
	for _, v := range g.RouteExclude {
		if v.Val != "" {
			ipMask, _, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("RouteExclude 错误" + err.Error())
			}
			v.IpMask = ipMask
			routeExclude = append(routeExclude, v)
		}
	}
	g.RouteExclude = routeExclude
	// 转换数据
	linkAcl := []GroupLinkAcl{}
	for _, v := range g.LinkAcl {
		if v.Val != "" {
			_, ipNet, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("GroupLinkAcl 错误" + err.Error())
			}
			v.IpNet = ipNet
			linkAcl = append(linkAcl, v)
		}
	}
	g.LinkAcl = linkAcl

	g.UpdatedAt = time.Now()

	if g.Id == 0 {
		g.CreatedAt = time.Now()
		err = Save(g)
	} else {
		err = Set("name", g.Name, g)
	}

	return err
}

func parseIpNet(s string) (string, *net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return "", nil, err
	}

	mask := net.IP(ipNet.Mask)
	ipMask := fmt.Sprintf("%s/%s", ip, mask)

	return ipMask, ipNet, nil
}
